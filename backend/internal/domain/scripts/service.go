package scripts

import (
	"context"
	"errors"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"

	"github.com/oklog/ulid/v2"
)

const serviceName = "scripts_service"

// Service is the script library's rules over its Store: who may read, change,
// delete and rate what. Every write runs in one transaction.
type Service struct {
	o11y   observability.Observer
	client database.Client
	store  Store
}

// ServiceOption configures a Service.
type ServiceOption func(*Service)

// WithServicePillars attaches the observability pillars.
func WithServicePillars(p *observability.Pillars) ServiceOption {
	return func(s *Service) {
		if p != nil {
			s.o11y = observability.NewObserver(serviceName, p.Logger, p.TracerProvider)
		}
	}
}

// NewService builds the service.
func NewService(client database.Client, store Store, opts ...ServiceOption) (*Service, error) {
	if client == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client for the scripts service")
	}

	if store == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil store for the scripts service")
	}

	s := &Service{
		o11y:   observability.NewObserver(serviceName, nil, nil),
		client: client,
		store:  store,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	return s, nil
}

// List is the scripts a viewer may see, with their own rating on each. An
// admin sees everything, private and disabled scripts included; anybody else
// sees public scripts and their own, and never a disabled one.
func (s *Service) List(ctx context.Context, viewer *Viewer, search string) ([]*Play, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	plays, err := s.store.ListPlays(ctx, s.client.Reader(), &ListFilter{
		ViewerID: viewer.UserID,
		Search:   search,
		All:      viewer.Admin,
	})
	if err != nil {
		return nil, op.Error(err, "listing scripts")
	}

	for _, play := range plays {
		if err = s.attachUserRating(ctx, viewer, play); err != nil {
			return nil, op.Error(err, "reading the viewer's ratings")
		}
	}

	return plays, nil
}

// Get is one script. A disabled script is ErrNotFound to anybody but an admin.
func (s *Service) Get(ctx context.Context, viewer *Viewer, id string) (*Play, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", id))
	defer op.End()

	play, err := s.store.GetPlay(ctx, s.client.Reader(), id)
	if err != nil {
		return nil, err
	}

	if !play.Enabled && !viewer.Admin {
		return nil, ErrNotFound
	}

	if err = s.attachUserRating(ctx, viewer, play); err != nil {
		return nil, op.Error(err, "reading the viewer's rating")
	}

	return play, nil
}

// Create writes a new script owned by the viewer. It starts enabled.
func (s *Service) Create(ctx context.Context, viewer *Viewer, input *PlayInput) (*Play, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	if viewer.UserID == "" {
		return nil, ErrNotSignedIn
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, err
	}

	id := ulid.Make().String()

	err := s.client.WithTransaction(ctx, func(tx database.Tx) error {
		resolved, resolveErr := s.resolveTags(ctx, tx, input)
		if resolveErr != nil {
			return resolveErr
		}

		return s.store.CreatePlay(ctx, tx, id, viewer.UserID, resolved)
	})
	if err != nil {
		return nil, s.refusal(op, err, "creating a script")
	}

	return s.Get(ctx, viewer, id)
}

// Update replaces a script the viewer owns. Anybody else's script is
// ErrNotFound, admins included: an admin's power is to delete and switch off,
// not to rewrite somebody's work.
func (s *Service) Update(ctx context.Context, viewer *Viewer, id string, input *PlayInput) (*Play, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", id))
	defer op.End()

	if viewer.UserID == "" {
		return nil, ErrNotSignedIn
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, err
	}

	err := s.client.WithTransaction(ctx, func(tx database.Tx) error {
		resolved, resolveErr := s.resolveTags(ctx, tx, input)
		if resolveErr != nil {
			return resolveErr
		}

		return s.store.ReplaceOwnedPlay(ctx, tx, viewer.UserID, id, resolved)
	})
	if err != nil {
		return nil, s.refusal(op, err, "updating a script")
	}

	return s.Get(ctx, viewer, id)
}

// Delete removes a script. It is admin-only, the owner included.
func (s *Service) Delete(ctx context.Context, viewer *Viewer, id string) error {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", id))
	defer op.End()

	if !viewer.Admin {
		return ErrNotAdmin
	}

	err := s.client.WithTransaction(ctx, func(tx database.Tx) error {
		return s.store.DeletePlay(ctx, tx, id)
	})
	if err != nil {
		return s.refusal(op, err, "deleting a script")
	}

	op.Logger().WithValue("identity.user_id", viewer.UserID).Info("admin deleted a script")

	return nil
}

// SetEnabled switches a script on or off. It is admin-only.
func (s *Service) SetEnabled(ctx context.Context, viewer *Viewer, id string, enabled bool) error {
	ctx, op := s.o11y.Begin(ctx,
		observability.WithValue("play_script.id", id),
		observability.WithValue("play_script.enabled", enabled),
	)
	defer op.End()

	if !viewer.Admin {
		return ErrNotAdmin
	}

	err := s.client.WithTransaction(ctx, func(tx database.Tx) error {
		return s.store.SetPlayEnabled(ctx, tx, id, enabled)
	})
	if err != nil {
		return s.refusal(op, err, "switching a script")
	}

	op.Logger().WithValue("identity.user_id", viewer.UserID).Info("admin switched a script")

	return nil
}

// Rate records the viewer's rating of a script, replacing any earlier one, and
// answers the script's ratings after it. Nobody rates their own script.
func (s *Service) Rate(ctx context.Context, viewer *Viewer, id string, rating int32) (*RatingStats, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", id))
	defer op.End()

	if viewer.UserID == "" {
		return nil, ErrNotSignedIn
	}

	if rating < MinRating || rating > MaxRating {
		return nil, ErrRatingOutOfRange
	}

	var stats *RatingStats

	err := s.client.WithTransaction(ctx, func(tx database.Tx) error {
		owner, ownerErr := s.store.PlayOwner(ctx, tx, id)
		if ownerErr != nil {
			return ownerErr
		}

		if owner == viewer.UserID {
			return ErrSelfRating
		}

		var rateErr error
		stats, rateErr = s.store.RatePlay(ctx, tx, viewer.UserID, id, rating)

		return rateErr
	})
	if err != nil {
		return nil, s.refusal(op, err, "rating a script")
	}

	return stats, nil
}

// Tags lists every tag.
func (s *Service) Tags(ctx context.Context) ([]*Tag, error) {
	return s.store.ListTags(ctx, s.client.Reader())
}

// resolveTags returns the input with its tag names turned into tag IDs,
// created where no tag has the name, and the IDs deduplicated.
func (s *Service) resolveTags(ctx context.Context, tx database.Tx, input *PlayInput) (*PlayInput, error) {
	resolved := *input
	resolved.TagIDs = nil

	seen := map[string]bool{}
	add := func(id string) {
		if !seen[id] {
			seen[id] = true
			resolved.TagIDs = append(resolved.TagIDs, id)
		}
	}

	for _, id := range input.TagIDs {
		add(id)
	}

	for _, name := range input.TagNames {
		if NormalizeTagName(name) == "" {
			continue
		}

		id, err := s.store.EnsureTag(ctx, tx, name)
		if err != nil {
			return nil, err
		}

		add(id)
	}

	return &resolved, nil
}

func (s *Service) attachUserRating(ctx context.Context, viewer *Viewer, play *Play) error {
	if viewer.UserID == "" {
		return nil
	}

	rating, rated, err := s.store.UserRating(ctx, s.client.Reader(), viewer.UserID, play.ID)
	if err != nil {
		return err
	}

	if rated {
		play.UserRating = new(rating)
	}

	return nil
}

// refusal hands back a refusal the caller can branch on as it is, and logs
// anything else as the failure it is.
func (s *Service) refusal(op observability.Operation, err error, description string) error {
	for _, refusal := range []error{ErrNotFound, ErrSelfRating, ErrInvalidInput, ErrNotAdmin, ErrNotSignedIn, ErrRatingOutOfRange} {
		if errors.Is(err, refusal) {
			return err
		}
	}

	return op.Error(err, "%s", description)
}
