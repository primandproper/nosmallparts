package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity/generated"
)

var (
	_ identity.AccountInvitationDataManager = (*repository)(nil)
)

// AccountInvitationExists fetches whether an account invitation exists from the database.
func (r *repository) AccountInvitationExists(ctx context.Context, accountInvitationID string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountInvitationID == "" {
		return false, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, keys.AccountInvitationIDKey, accountInvitationID)

	result, err := r.generatedQuerier.CheckAccountInvitationExistence(ctx, r.db, accountInvitationID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing account invitation existence check")
	}

	return result, nil
}

// GetAccountInvitationByAccountAndID fetches an invitation from the database.
func (r *repository) GetAccountInvitationByAccountAndID(ctx context.Context, accountID, accountInvitationID string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	if accountInvitationID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, keys.AccountInvitationIDKey, accountInvitationID)

	result, err := r.generatedQuerier.GetAccountInvitationByAccountAndID(ctx, r.db, &generated.GetAccountInvitationByAccountAndIDParams{
		DestinationAccount: accountID,
		ID:                 accountInvitationID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	accountInvitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			AvatarSrc:                  database.StringPointerFromNullString(result.UserAvatarSrc),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			ServiceRole:                result.UserServiceRole,
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return accountInvitation, nil
}

// GetAccountInvitationByTokenAndID fetches an invitation from the database.
func (r *repository) GetAccountInvitationByTokenAndID(ctx context.Context, token, invitationID string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if token == "" {
		return nil, database.ErrInvalidIDProvided
	}

	if invitationID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountInvitationIDKey, invitationID)
	tracing.AttachToSpan(span, keys.AccountInvitationIDKey, invitationID)

	logger.Debug("fetching account invitation")

	result, err := r.generatedQuerier.GetAccountInvitationByTokenAndID(ctx, r.db, &generated.GetAccountInvitationByTokenAndIDParams{
		Token: token,
		ID:    invitationID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	accountInvitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			AvatarSrc:                  database.StringPointerFromNullString(result.UserAvatarSrc),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			ServiceRole:                result.UserServiceRole,
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return accountInvitation, nil
}

// GetAccountInvitationByToken fetches an invitation from the database.
func (r *repository) GetAccountInvitationByToken(ctx context.Context, token string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if token == "" {
		return nil, database.ErrInvalidIDProvided
	}

	logger.Debug("fetching account invitation")

	result, err := r.generatedQuerier.GetAccountInvitationByToken(ctx, r.db, token)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	accountInvitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			AvatarSrc:                  database.StringPointerFromNullString(result.UserAvatarSrc),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			ServiceRole:                result.UserServiceRole,
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return accountInvitation, nil
}

// GetAccountInvitationByEmailAndToken fetches an invitation from the database.
func (r *repository) GetAccountInvitationByEmailAndToken(ctx context.Context, emailAddress, token string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if emailAddress == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.UserEmailAddressKey, emailAddress)
	tracing.AttachToSpan(span, keys.UserEmailAddressKey, emailAddress)

	if token == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountInvitationTokenKey, token)
	tracing.AttachToSpan(span, keys.AccountInvitationTokenKey, token)

	result, err := r.generatedQuerier.GetAccountInvitationByEmailAndToken(ctx, r.db, &generated.GetAccountInvitationByEmailAndTokenParams{
		ToEmail: emailAddress,
		Token:   token,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	invitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			AvatarSrc:                  database.StringPointerFromNullString(result.UserAvatarSrc),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			ServiceRole:                result.UserServiceRole,
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return invitation, nil
}

// CreateAccountInvitation creates an invitation in a database.
func (r *repository) CreateAccountInvitation(ctx context.Context, input *identity.AccountInvitationDatabaseCreationInput) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}

	logger := r.logger.WithValue(keys.AccountInvitationIDKey, input.ID)
	tracing.AttachToSpan(span, keys.AccountIDKey, input.DestinationAccountID)

	if input.ToUser == nil && input.ToEmail != "" {
		if invitee, err := r.GetUserByEmail(ctx, input.ToEmail); err == nil {
			input.ToUser = &invitee.ID
		}
	}

	if err := r.generatedQuerier.CreateAccountInvitation(ctx, r.db, &generated.CreateAccountInvitationParams{
		ExpiresAt:          input.ExpiresAt,
		ID:                 input.ID,
		FromUser:           input.FromUser,
		ToName:             input.ToName,
		Note:               input.Note,
		ToEmail:            input.ToEmail,
		Token:              input.Token,
		DestinationAccount: input.DestinationAccountID,
		ToUser:             database.NullStringFromStringPointer(input.ToUser),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing account invitation creation query")
	}

	x := &identity.AccountInvitation{
		ID:                 input.ID,
		FromUser:           identity.User{ID: input.FromUser},
		ToUser:             input.ToUser,
		Note:               input.Note,
		ToName:             input.ToName,
		ToEmail:            input.ToEmail,
		Token:              input.Token,
		StatusNote:         "",
		Status:             string(identity.PendingAccountInvitationStatus),
		DestinationAccount: identity.Account{ID: input.DestinationAccountID},
		ExpiresAt:          input.ExpiresAt,
		CreatedAt:          r.CurrentTime(),
	}

	tracing.AttachToSpan(span, keys.AccountInvitationIDKey, x.ID)
	logger = logger.WithValue(keys.AccountInvitationIDKey, x.ID)

	logger.Info("account invitation created")

	return x, nil
}

// GetPendingAccountInvitationsFromUser fetches pending account invitations sent from a given user.
func (r *repository) GetPendingAccountInvitationsFromUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger := r.logger.WithValue(keys.UserIDKey, userID)
	filter.AttachToLogger(logger)

	results, err := r.generatedQuerier.GetPendingInvitesFromUser(ctx, r.db, &generated.GetPendingInvitesFromUserParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Status:          generated.InvitationState(identity.PendingAccountInvitationStatus),
		FromUser:        userID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing account invitation query")
	}

	var (
		data                      []*identity.AccountInvitation
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &identity.AccountInvitation{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ToUser:        database.StringPointerFromNullString(result.ToUser),
			Status:        string(result.Status),
			ToEmail:       result.ToEmail,
			StatusNote:    result.StatusNote,
			Token:         result.Token,
			ID:            result.ID,
			Note:          result.Note,
			ToName:        result.ToName,
			ExpiresAt:     result.ExpiresAt,
			DestinationAccount: identity.Account{
				CreatedAt:                  result.AccountCreatedAt,
				SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
				LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
				ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
				ContactPhone:               result.AccountContactPhone,
				BillingStatus:              result.AccountBillingStatus,
				AddressLine1:               result.AccountAddressLine1,
				AddressLine2:               result.AccountAddressLine2,
				City:                       result.AccountCity,
				State:                      result.AccountState,
				ZipCode:                    result.AccountZipCode,
				Country:                    result.AccountCountry,
				Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
				Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
				PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
				BelongsToUser:              result.AccountBelongsToUser,
				ID:                         result.AccountID,
				Name:                       result.AccountName,
				Members:                    nil,
			},
			FromUser: identity.User{
				CreatedAt:                  result.UserCreatedAt,
				PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
				LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
				LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
				LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
				TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
				AvatarSrc:                  database.StringPointerFromNullString(result.UserAvatarSrc),
				Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
				ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
				AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
				TwoFactorSecret:            result.UserTwoFactorSecret,
				HashedPassword:             result.UserHashedPassword,
				ID:                         result.UserID,
				AccountStatus:              result.UserUserAccountStatus,
				Username:                   result.UserUsername,
				FirstName:                  result.UserFirstName,
				LastName:                   result.UserLastName,
				EmailAddress:               result.UserEmailAddress,
				EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
				ServiceRole:                result.UserServiceRole,
				RequiresPasswordChange:     result.UserRequiresPasswordChange,
			},
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *identity.AccountInvitation) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetPendingAccountInvitationsForUser fetches pending account invitations sent to a given user.
func (r *repository) GetPendingAccountInvitationsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger := r.logger.WithValue(keys.UserIDKey, userID)
	filter.AttachToLogger(logger)

	results, err := r.generatedQuerier.GetPendingInvitesForUser(ctx, r.db, &generated.GetPendingInvitesForUserParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Status:          generated.InvitationState(identity.PendingAccountInvitationStatus),
		ToUser:          database.NullStringFromString(userID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing account invitation query")
	}

	var (
		data                      []*identity.AccountInvitation
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &identity.AccountInvitation{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ToUser:        database.StringPointerFromNullString(result.ToUser),
			Status:        string(result.Status),
			ToEmail:       result.ToEmail,
			StatusNote:    result.StatusNote,
			Token:         result.Token,
			ID:            result.ID,
			Note:          result.Note,
			ToName:        result.ToName,
			ExpiresAt:     result.ExpiresAt,
			DestinationAccount: identity.Account{
				CreatedAt:                  result.AccountCreatedAt,
				SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
				LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
				ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
				ContactPhone:               result.AccountContactPhone,
				BillingStatus:              result.AccountBillingStatus,
				AddressLine1:               result.AccountAddressLine1,
				AddressLine2:               result.AccountAddressLine2,
				City:                       result.AccountCity,
				State:                      result.AccountState,
				ZipCode:                    result.AccountZipCode,
				Country:                    result.AccountCountry,
				Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
				Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
				PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
				BelongsToUser:              result.AccountBelongsToUser,
				ID:                         result.AccountID,
				Name:                       result.AccountName,
				Members:                    nil,
			},
			FromUser: identity.User{
				CreatedAt:                  result.UserCreatedAt,
				PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
				LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
				LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
				LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
				TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
				AvatarSrc:                  database.StringPointerFromNullString(result.UserAvatarSrc),
				Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
				ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
				AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
				TwoFactorSecret:            result.UserTwoFactorSecret,
				HashedPassword:             result.UserHashedPassword,
				ID:                         result.UserID,
				AccountStatus:              result.UserUserAccountStatus,
				Username:                   result.UserUsername,
				FirstName:                  result.UserFirstName,
				LastName:                   result.UserLastName,
				EmailAddress:               result.UserEmailAddress,
				EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
				ServiceRole:                result.UserServiceRole,
				RequiresPasswordChange:     result.UserRequiresPasswordChange,
			},
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *identity.AccountInvitation) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

func (r *repository) setInvitationStatus(ctx context.Context, querier database.SQLQueryExecutor, accountInvitationID, note, status string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithValue("new_status", status)

	if accountInvitationID == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, keys.AccountInvitationIDKey, accountInvitationID)

	if err := r.generatedQuerier.SetAccountInvitationStatus(ctx, querier, &generated.SetAccountInvitationStatusParams{
		Status:     generated.InvitationState(status),
		StatusNote: note,
		ID:         accountInvitationID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "changing account invitation status")
	}

	logger.Debug("account invitation updated")

	return nil
}

// CancelAccountInvitation cancels an account invitation by its MealPlanTaskID with a note.
func (r *repository) CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	return r.setInvitationStatus(ctx, r.db, accountInvitationID, note, string(identity.CancelledAccountInvitationStatus))
}

// AcceptAccountInvitation accepts an account invitation by its MealPlanTaskID with a note.
func (r *repository) AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID, token, note string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountInvitationID == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, keys.AccountInvitationIDKey, accountInvitationID)

	if token == "" {
		return database.ErrNilInputProvided
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	invitation, err := r.GetAccountInvitationByTokenAndID(ctx, token, accountInvitationID)
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	if err = r.setInvitationStatus(ctx, tx, accountInvitationID, note, string(identity.AcceptedAccountInvitationStatus)); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "accepting account invitation")
	}

	addUserInput := &identity.AccountUserMembershipDatabaseCreationInput{
		ID:          identifiers.New(),
		Reason:      fmt.Sprintf("accepted account invitation %s", accountInvitationID),
		AccountID:   invitation.DestinationAccount.ID,
		AccountRole: "account_member",
	}
	if invitation.ToUser != nil {
		addUserInput.UserID = *invitation.ToUser
		if err = r.addUserToAccount(ctx, tx, addUserInput); err != nil {
			r.RollbackTransaction(ctx, tx)
			return observability.PrepareAndLogError(err, logger, span, "adding user to account")
		}
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	return nil
}

// RejectAccountInvitation rejects an account invitation by its MealPlanTaskID with a note.
func (r *repository) RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	return r.setInvitationStatus(ctx, r.db, accountInvitationID, note, string(identity.RejectedAccountInvitationStatus))
}

func (r *repository) attachInvitationsToUser(ctx context.Context, querier database.SQLQueryExecutor, userEmail, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger

	if userEmail == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.UserEmailAddressKey, userEmail)
	tracing.AttachToSpan(span, keys.AccountIDKey, userEmail)

	if userID == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.AccountInvitationIDKey, userID)

	rowCount, err := r.generatedQuerier.AttachAccountInvitationsToUserID(ctx, querier, &generated.AttachAccountInvitationsToUserIDParams{
		ToEmail: userEmail,
		ToUser:  database.NullStringFromString(userID),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareAndLogError(err, logger, span, "attaching invitations to user")
	}

	logger.WithValue("rows_affected", rowCount).Info("invitations associated with user")

	return nil
}

func (r *repository) acceptInvitationForUser(ctx context.Context, querier database.SQLQueryExecutorAndTransactionManager, input *identity.UserDatabaseCreationInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithValue(keys.UsernameKey, input.Username).WithValue(keys.UserEmailAddressKey, input.EmailAddress)

	invitation, tokenCheckErr := r.GetAccountInvitationByToken(ctx, input.InvitationToken)
	if tokenCheckErr != nil {
		r.RollbackTransaction(ctx, querier)
		return observability.PrepareError(tokenCheckErr, span, "fetching account invitation")
	}

	if err := r.generatedQuerier.CreateAccountUserMembershipForNewUser(ctx, querier, &generated.CreateAccountUserMembershipForNewUserParams{
		ID:               identifiers.New(),
		BelongsToUser:    input.ID,
		BelongsToAccount: input.DestinationAccountID,
		AccountRole:      authorization.AccountMemberRole.String(),
		DefaultAccount:   true,
	}); err != nil {
		r.RollbackTransaction(ctx, querier)
		return observability.PrepareAndLogError(err, logger, span, "writing destination account membership")
	}

	logger.Debug("created membership via invitation")

	if err := r.setInvitationStatus(ctx, querier, invitation.ID, "", string(identity.AcceptedAccountInvitationStatus)); err != nil {
		r.RollbackTransaction(ctx, querier)
		return observability.PrepareAndLogError(err, logger, span, "accepting account invitation")
	}

	logger.Debug("marked invitation as accepted")

	return nil
}
