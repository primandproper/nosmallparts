package manager

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/converters"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	platformkeys "github.com/verygoodsoftwarenotvirus/platform/v4/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/random"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	todoRes  = "TODO"
	o11yName = "identity_data_manager"

	totpIssuer             = "DinnerDoneBetter"
	base64ImagePrefix      = "data:image/jpeg;base64,"
	minimumPasswordEntropy = 60
	totpSecretSize         = 64
)

var (
	userAvatarBase64Encoding = base64.RawURLEncoding

	// ErrInvalidIDProvided indicates a required ID was passed in empty.
	ErrInvalidIDProvided = errors.New("required ID was empty")

	// ErrNilInputProvided indicates that a required parameter was nil.
	ErrNilInputProvided = errors.New("nil input provided")

	// ErrEmptyInputProvided indicates that the required input was empty.
	ErrEmptyInputProvided = errors.New("empty input provided")
)

type (
	manager struct {
		tracer               tracing.Tracer
		logger               logging.Logger
		dataChangesPublisher messagequeue.Publisher
		identityRepo         identity.Repository
		secretGenerator      random.Generator
		authenticator        authentication.Hasher
		userSearchIndex      indexing.UserTextSearcher
	}
)

func NewIdentityDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	publisherProvider messagequeue.PublisherProvider,
	identityRepo identity.Repository,
	secretGenerator random.Generator,
	authenticator authentication.Hasher,
	userSearchIndex indexing.UserTextSearcher,
	cfg *msgconfig.QueuesConfig,
) (IdentityDataManager, error) {
	publisher, err := publisherProvider.ProvidePublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("setting up data changes publisher: %w", err)
	}

	return &manager{
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:               logging.NewNamedLogger(logger, o11yName),
		identityRepo:         identityRepo,
		dataChangesPublisher: publisher,
		secretGenerator:      secretGenerator,
		authenticator:        authenticator,
		userSearchIndex:      userSearchIndex,
	}, nil
}

func (m *manager) AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountInvitationID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	if err := m.identityRepo.AcceptAccountInvitation(ctx, accountID, accountInvitationID, input.Token, input.Note); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "accepting account invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationCanceledServiceEventType, map[string]any{
		keys.AccountInvitationIDKey: accountInvitationID,
	}))

	return nil
}

func (m *manager) RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountInvitationID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	// note, this is where you would call input.ValidateWithContext, if that currently had any effect.

	invitation, err := m.identityRepo.GetAccountInvitationByTokenAndID(ctx, input.Token, accountInvitationID)
	if errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "account invitation not found")
	} else if err != nil {
		return observability.PrepareError(err, span, "retrieving invitation")
	}

	if err = m.identityRepo.RejectAccountInvitation(ctx, accountID, invitation.ID, input.Note); err != nil {
		return observability.PrepareError(err, span, "rejecting invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationRejectedServiceEventType, map[string]any{
		keys.AccountInvitationIDKey: accountInvitationID,
	}))

	return nil
}

func (m *manager) CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountInvitationID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	if err := m.identityRepo.CancelAccountInvitation(ctx, accountID, accountInvitationID, note); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "canceling account invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationCanceledServiceEventType, map[string]any{
		keys.AccountInvitationIDKey: accountInvitationID,
	}))

	return nil
}

func (m *manager) ArchiveAccount(ctx context.Context, accountID, ownerID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || ownerID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountIDKey: accountID,
		platformkeys.UserIDKey:    ownerID,
	}, span, m.logger)

	if err := m.identityRepo.ArchiveAccount(ctx, accountID, ownerID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "ArchiveAccount")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountArchivedServiceEventType, map[string]any{
		keys.AccountIDKey: accountID,
		platformkeys.UserIDKey:    ownerID,
	}))

	return nil
}

func (m *manager) ArchiveUserMembership(ctx context.Context, userID, accountID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || userID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountIDKey: accountID,
		platformkeys.UserIDKey:    userID,
	}, span, m.logger)

	if err := m.identityRepo.RemoveUserFromAccount(ctx, userID, accountID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "RemoveUserFromAccount")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountMemberRemovedServiceEventType, map[string]any{
		keys.AccountIDKey: accountID,
		platformkeys.UserIDKey:    userID,
	}))

	return nil
}

func (m *manager) ArchiveUser(ctx context.Context, userID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.ArchiveUser(ctx, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "ArchiveUser")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserArchivedServiceEventType, map[string]any{
		platformkeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) CreateAccount(ctx context.Context, input *identity.AccountCreationRequestInput) (*identity.Account, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := m.logger.WithSpan(span)

	created, err := m.identityRepo.CreateAccount(ctx, converters.ConvertAccountCreationInputToAccountDatabaseCreationInput(input))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating Account")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountCreatedServiceEventType, map[string]any{
		keys.AccountIDKey: created.ID,
	}))

	return created, nil
}

func (m *manager) CreateAccountInvitation(ctx context.Context, userID, accountID string, input *identity.AccountInvitationCreationRequestInput) (*identity.AccountInvitation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	}, span, m.logger)

	token, err := m.secretGenerator.GenerateBase64EncodedString(ctx, 64)
	if err != nil {
		return nil, observability.PrepareError(err, span, "generating account invitation token")
	}

	convertedInput := converters.ConvertAccountInvitationCreationInputToAccountInvitationDatabaseCreationInput(userID, accountID, token, input)

	created, err := m.identityRepo.CreateAccountInvitation(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating account invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationCreatedServiceEventType, map[string]any{
		keys.AccountInvitationIDKey: created.ID,
		platformkeys.UserIDKey:              userID,
		"destination_account":       accountID,
	}))

	return created, nil
}

func (m *manager) CreateUser(ctx context.Context, input *identity.UserRegistrationInput) (*identity.UserCreationResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UsernameKey: input.Username,
	}, span, m.logger)

	input.Username = strings.TrimSpace(input.Username)
	tracing.AttachToSpan(span, platformkeys.UsernameKey, input.Username)
	input.EmailAddress = strings.TrimSpace(strings.ToLower(input.EmailAddress))
	tracing.AttachToSpan(span, keys.UserEmailAddressKey, input.EmailAddress)
	input.Password = strings.TrimSpace(input.Password)

	logger = logger.WithValues(map[string]any{
		platformkeys.UsernameKey:               input.Username,
		keys.UserEmailAddressKey:       input.EmailAddress,
		keys.AccountInvitationIDKey:    input.InvitationID,
		keys.AccountInvitationTokenKey: input.InvitationToken,
	})

	if err := input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(platformkeys.ValidationErrorKey, err).Debug("provided dbInput was invalid")
		return nil, observability.PrepareError(err, span, "invalid user creation dbInput provided")
	}

	// ensure the password is not garbage-tier
	if err := passwordvalidator.Validate(strings.TrimSpace(input.Password), minimumPasswordEntropy); err != nil {
		logger.WithValue("password_validation_error", err).Debug("weak password provided to user creation route")
		return nil, observability.PrepareAndLogError(err, logger, span, "weak password provided for user creation")
	}

	var invitation *identity.AccountInvitation
	if input.InvitationID != "" && input.InvitationToken != "" {
		invite, err := m.identityRepo.GetAccountInvitationByTokenAndID(ctx, input.InvitationToken, input.InvitationID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, observability.PrepareAndLogError(err, logger, span, "no account invitation found")
		} else if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "getting account invitation by token and ID")
		}

		invitation = invite
		logger.Debug("retrieved account invitation")
	}

	logger.Debug("completed invitation check")

	// hash the password
	hp, err := m.authenticator.HashPassword(ctx, strings.TrimSpace(input.Password))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "hashing user creation password")
	}

	// generate a two-factor secret.
	tfs, err := m.secretGenerator.GenerateBase32EncodedString(ctx, totpSecretSize)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "generating two factor secret")
	}

	dbInput := &identity.UserDatabaseCreationInput{
		ID:              identifiers.New(),
		Username:        input.Username,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		EmailAddress:    input.EmailAddress,
		HashedPassword:  hp,
		TwoFactorSecret: tfs,
		InvitationToken: input.InvitationToken,
		Birthday:        input.Birthday,
		AccountName:     input.AccountName,
	}

	if invitation != nil {
		logger.Debug("supplementing user creation dbInput with invitation data")
		dbInput.DestinationAccountID = invitation.DestinationAccount.ID
		dbInput.InvitationToken = invitation.Token
	}

	// create the user.
	user, err := m.identityRepo.CreateUser(ctx, dbInput)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating user")
		if errors.Is(err, database.ErrUserAlreadyExists) {
			return nil, observability.PrepareAndLogError(err, logger, span, "user already exists")
		}
		return nil, observability.PrepareAndLogError(err, logger, span, "creating user in database")
	}

	logger.Debug("user created")

	defaultAccountID, err := m.identityRepo.GetDefaultAccountIDForUser(ctx, user.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching default account ID for user")
	}

	emailVerificationToken, err := m.identityRepo.GetEmailAddressVerificationTokenForUser(ctx, user.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching email verification token for user")
	}

	// notify the relevant parties.
	tracing.AttachToSpan(span, platformkeys.UserIDKey, user.ID)

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserSignedUpServiceEventType, map[string]any{
		keys.AccountIDKey:                  defaultAccountID,
		platformkeys.UserIDKey:                     user.ID,
		keys.UserEmailVerificationTokenKey: emailVerificationToken,
	}))

	/* TODO: this should be done in a downstream handler

	if err = m.analyticsReporter.AddUser(ctx, user.ID, map[string]any{
		"username":        user.Username,
		"default_account": defaultAccountID,
		"first_name":      user.FirstName,
		"last_name":       user.LastName,
	}); err != nil {
		observability.AcknowledgeError(err, logger, span, "identifying user for analytics")
	}

	if err = s.featureFlagManager.Identify(ctx, user); err != nil {
		observability.AcknowledgeError(err, logger, span, "identifying user in feature flag manager")
	}
	*/

	// UserCreationResponse is a struct we can use to notify the user of their two factor secret, but ideally just this once and then never again.
	ucr := &identity.UserCreationResponse{
		CreatedUserID:   user.ID,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		EmailAddress:    user.EmailAddress,
		CreatedAt:       user.CreatedAt,
		TwoFactorSecret: user.TwoFactorSecret,
		Birthday:        user.Birthday,
		TwoFactorQRCode: m.buildQRCode(ctx, user.Username, user.TwoFactorSecret),
	}

	return ucr, nil
}

func (m *manager) GetAccount(ctx context.Context, accountID string) (*identity.Account, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountIDKey: accountID,
	}, span, m.logger)

	account, err := m.identityRepo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account")
	}

	return account, nil
}

func (m *manager) GetAccountInvitation(ctx context.Context, accountID, accountInvitationID string) (*identity.AccountInvitation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || accountInvitationID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountIDKey:           accountID,
		keys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	invitation, err := m.identityRepo.GetAccountInvitationByAccountAndID(ctx, accountID, accountInvitationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting invitation")
	}

	return invitation, nil
}

func (m *manager) GetAccounts(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	accounts, err := m.identityRepo.GetAccounts(ctx, userID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching accounts")
	}

	return accounts, nil
}

func (m *manager) GetReceivedAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	invites, err := m.identityRepo.GetPendingAccountInvitationsForUser(ctx, userID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching invites")
	}

	return invites, nil
}

func (m *manager) GetSentAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	invites, err := m.identityRepo.GetPendingAccountInvitationsFromUser(ctx, userID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching invites")
	}

	return invites, nil
}

func (m *manager) GetUser(ctx context.Context, userID string) (*identity.User, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	user, err := m.identityRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting user")
	}

	return user, nil
}

func (m *manager) GetUsers(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span)

	users, err := m.identityRepo.GetUsers(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching users")
	}

	return users, nil
}

func (m *manager) GetUsersForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span)

	users, err := m.identityRepo.GetUsersForAccount(ctx, accountID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching users")
	}

	return users, nil
}

func (m *manager) SearchForUsers(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UseDatabaseKey: !useSearchService,
	}, span, m.logger)

	if !useSearchService {
		users, err := m.identityRepo.SearchForUsersByUsername(ctx, query, filter)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, observability.PrepareError(err, span, "no users found")
			}
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for users")
		}

		return users, nil
	} else {
		uss, err := m.userSearchIndex.Search(ctx, query)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for users")
		}

		userIDs := []string{}
		for _, us := range uss {
			userIDs = append(userIDs, us.ID)
		}

		users, err := m.identityRepo.GetUsersWithIDs(ctx, userIDs)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for users")
		}

		result := filtering.NewQueryFilteredResult(users, uint64(len(users)), uint64(len(users)), func(u *identity.User) string {
			return u.ID
		}, filter)

		return result, nil
	}
}

func (m *manager) SetDefaultAccount(ctx context.Context, userID, accountID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	}, span, m.logger)

	// mark household as default in database.
	if err := m.identityRepo.MarkAccountAsUserDefault(ctx, userID, accountID); err != nil {
		return observability.PrepareError(err, span, "marking default account as user")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountSetAsDefaultServiceEventType, map[string]any{
		keys.AccountIDKey: accountID,
		platformkeys.UserIDKey:    userID,
	}))

	return nil
}

func (m *manager) TransferAccountOwnership(ctx context.Context, accountID string, input *identity.AccountOwnershipTransferInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountIDKey: accountID,
	}, span, m.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "")
	}

	// transfer ownership of household in database.
	if err := m.identityRepo.TransferAccountOwnership(ctx, accountID, input); err != nil {
		return observability.PrepareError(err, span, "")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountOwnershipTransferredServiceEventType, map[string]any{
		keys.AccountIDKey: accountID,
	}))

	return nil
}

func (m *manager) UpdateAccount(ctx context.Context, accountID string, input *identity.AccountUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		keys.AccountIDKey: accountID,
	}, span, m.logger)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "validating account update")
	}

	// fetch the account from the database.
	account, err := m.identityRepo.GetAccount(ctx, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "no account found")
	} else if err != nil {
		return observability.PrepareError(err, span, "fetching account")
	}

	account.Update(input)

	// update the account in the database.
	if err = m.identityRepo.UpdateAccount(ctx, account); err != nil {
		return observability.PrepareError(err, span, "updating account")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountUpdatedServiceEventType, map[string]any{
		keys.AccountIDKey: accountID,
	}))

	return nil
}

func (m *manager) UpdateAccountMemberPermissions(ctx context.Context, accountID, userID string, input *identity.ModifyUserPermissionsInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	if err := m.identityRepo.ModifyUserPermissions(ctx, accountID, userID, input); err != nil {
		return observability.PrepareError(err, span, "modifying user permissions")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountMembershipPermissionsUpdatedServiceEventType, map[string]any{
		keys.AccountIDKey: accountID,
	}))

	return nil
}

func (m *manager) UpdateUserDetails(ctx context.Context, userID string, input *identity.UserDetailsUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserDetails(ctx, userID, converters.ConvertUserDetailsUpdateRequestInputToUserDetailsUpdateInput(input)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user details")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserDetailsChangedEventType, map[string]any{
		platformkeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) UpdateUserEmailAddress(ctx context.Context, userID, newEmail string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserEmailAddress(ctx, userID, newEmail); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user email address")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.EmailAddressChangedEventType, map[string]any{
		platformkeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) UpdateUserUsername(ctx context.Context, userID, newUsername string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	if newUsername == "" {
		return ErrEmptyInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserUsername(ctx, userID, newUsername); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user username")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UsernameChangedEventType, map[string]any{
		platformkeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) UploadUserAvatar(ctx context.Context, userID, base64EncodedImageData string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: userID,
	}, span, m.logger)

	data, err := userAvatarBase64Encoding.DecodeString(base64EncodedImageData)
	if err != nil {
		return observability.PrepareError(err, span, "decoding base64 encoded image data")
	}

	logger = observability.ObserveValues(map[string]any{
		"data.length": len(data),
	}, span, logger)

	if err = m.identityRepo.UpdateUserAvatar(ctx, userID, base64EncodedImageData); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user avatar")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserAvatarChangedEventType, map[string]any{
		platformkeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) AdminUpdateUserStatus(ctx context.Context, input *identity.UserAccountStatusUpdateInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UserIDKey: input.TargetUserID,
		platformkeys.ReasonKey: input.Reason,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserAccountStatus(ctx, input.TargetUserID, input); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user account status")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserStatusChangedServiceEventType, map[string]any{
		platformkeys.UserIDKey: input.TargetUserID,
	}))

	return nil
}

// buildQRCode builds a QR code for a given username and secret.
func (m *manager) buildQRCode(ctx context.Context, username, twoFactorSecret string) string {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UsernameKey: username,
	}, span, m.logger)

	// "otpauth://totp/{{ .Issuer }}:{{ .EnsureUsername }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
	otpString := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		totpIssuer,
		username,
		twoFactorSecret,
		totpIssuer,
	)

	// encode two factor secret as authenticator-friendly QR code
	qrCode, err := qr.Encode(otpString, qr.L, qr.Auto)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding OTP string")
		return ""
	}

	// scale the QR code so that it's not a PNG for ants.
	qrCode, err = barcode.Scale(qrCode, 256, 256)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "scaling QR code")
		return ""
	}

	// encode the QR code to PNG.
	var b bytes.Buffer
	if err = png.Encode(&b, qrCode); err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding QR code to PNG")
		return ""
	}

	// base64 encode the image for easy HTML use.
	return fmt.Sprintf("%s%s", base64ImagePrefix, base64.StdEncoding.EncodeToString(b.Bytes()))
}
