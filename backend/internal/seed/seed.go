// Package seed puts the built-in sample plays into a database, as public
// scripts owned by a system user nobody can sign in as.
//
// It is idempotent: the system user is found by its username and each play by
// its stable ID, and anything already there is left alone, so running it twice,
// or against a database that has had its sample plays edited, changes nothing.
package seed

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"

	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"

	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// The system user the sample plays belong to. It holds no password, so no
// sign-in door admits it.
const (
	SystemUsername    = "system@nosmallparts.local"
	systemDisplayName = "No Small Parts"
)

const seederName = "seeder"

//go:embed sample_plays.json
var samplePlaysJSON []byte

type (
	// samplePlay is a play as sample_plays.json spells it, which is the shape
	// the TypeScript server's sample plays had.
	samplePlay struct {
		Author      *string            `json:"author"`
		ID          string             `json:"id"`
		Title       string             `json:"title"`
		Description string             `json:"description"`
		Cast        []*sampleCastEntry `json:"cast"`
		Lines       []*sampleLine      `json:"lines"`
	}

	sampleCastEntry struct {
		Description *string `json:"description"`
		ID          string  `json:"id"`
		Name        string  `json:"name"`
	}

	sampleLine struct {
		Direction     *string            `json:"direction"`
		AutoScrollWPM *int32             `json:"autoScrollWpm"`
		Text          string             `json:"text"`
		ScrollMode    scripts.ScrollMode `json:"scrollMode"`
		CastMemberIDs []string           `json:"castMemberIds"`
		Index         int32              `json:"index"`
	}
)

// Seeder writes the sample plays.
type Seeder struct {
	o11y       observability.Observer
	client     database.Client
	directory  identity.Store
	registrar  *identity.Service
	store      scripts.Store
	ownerRoles []string
}

// NewSeeder builds a seeder. ownerRoles are what the system user holds in the
// account registration mints for it, which are sign-in's default owner roles.
func NewSeeder(
	client database.Client,
	directory identity.Store,
	registrar *identity.Service,
	store scripts.Store,
	ownerRoles []string,
	pillars *observability.Pillars,
) (*Seeder, error) {
	if client == nil || directory == nil || registrar == nil || store == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "building the seeder")
	}

	if len(ownerRoles) == 0 {
		return nil, platformerrors.New("the seeder needs the owner roles a registration mints")
	}

	o11y := observability.NewObserver(seederName, nil, nil)
	if pillars != nil {
		o11y = observability.NewObserver(seederName, pillars.Logger, pillars.TracerProvider)
	}

	return &Seeder{
		o11y:       o11y,
		client:     client,
		directory:  directory,
		registrar:  registrar,
		store:      store,
		ownerRoles: ownerRoles,
	}, nil
}

// Seed writes the system user and every sample play that is not already there.
func (s *Seeder) Seed(ctx context.Context) error {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	plays, err := samplePlays()
	if err != nil {
		return op.Error(err, "reading the sample plays")
	}

	ownerID, err := s.ensureSystemUser(ctx)
	if err != nil {
		return op.Error(err, "ensuring the system user")
	}

	seeded := 0

	for _, play := range plays {
		_, err = s.store.PlayOwner(ctx, s.client.Reader(), play.ID)
		if err == nil {
			continue
		}

		if !errors.Is(err, scripts.ErrNotFound) {
			return op.Error(err, "looking for sample play %s", play.ID)
		}

		if err = s.client.WithTransaction(ctx, func(tx database.Tx) error {
			return s.store.CreatePlay(ctx, tx, play.ID, ownerID, play.input())
		}); err != nil {
			return op.Error(err, "seeding sample play %s", play.ID)
		}

		seeded++

		op.Logger().WithValue("play_script.id", play.ID).Info("seeded a sample play")
	}

	op.Logger().WithValue("seeded", seeded).WithValue("sample_plays", len(plays)).Info("seed complete")

	return nil
}

// ensureSystemUser finds the system user, registering it the first time.
func (s *Seeder) ensureSystemUser(ctx context.Context) (string, error) {
	existing, err := s.directory.GetUserByUsername(ctx, s.client.Reader(), tenancy.Global(), SystemUsername)
	if err == nil {
		return existing.ID, nil
	}

	if !errors.Is(err, identity.ErrUserNotFound) {
		return "", platformerrors.Wrap(err, "looking up the system user")
	}

	registration, err := s.registrar.Register(ctx, tenancy.Global(),
		&identity.User{
			Username:     SystemUsername,
			EmailAddress: SystemUsername,
			DisplayName:  systemDisplayName,
		},
		&identity.Account{Name: systemDisplayName},
		s.ownerRoles,
	)
	if err != nil {
		return "", platformerrors.Wrap(err, "registering the system user")
	}

	return registration.User.ID, nil
}

func samplePlays() ([]*samplePlay, error) {
	var plays []*samplePlay
	if err := json.Unmarshal(samplePlaysJSON, &plays); err != nil {
		return nil, platformerrors.Wrap(err, "decoding sample_plays.json")
	}

	return plays, nil
}

// input is the play as an author would have written it. Sample plays are
// public.
func (p *samplePlay) input() *scripts.PlayInput {
	in := &scripts.PlayInput{
		Title:       p.Title,
		Description: p.Description,
		Author:      p.Author,
		IsPublic:    true,
	}

	for _, member := range p.Cast {
		in.Cast = append(in.Cast, &scripts.CastMemberInput{ID: member.ID, Name: member.Name, Description: member.Description})
	}

	for _, line := range p.Lines {
		in.Lines = append(in.Lines, &scripts.LineInput{
			CastMemberIDs: line.CastMemberIDs,
			Text:          line.Text,
			Index:         line.Index,
			Direction:     line.Direction,
			ScrollMode:    line.ScrollMode,
			AutoScrollWPM: line.AutoScrollWPM,
		})
	}

	return in
}
