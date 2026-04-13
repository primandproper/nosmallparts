package identity

import (
	"context"
	"database/sql"
	"errors"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	platformkeys "github.com/verygoodsoftwarenotvirus/platform/v4/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity/generated"
)

const (
	resourceTypeAccountUserMemberships = "account_user_memberships"
)

var (
	_ identity.AccountUserMembershipDataManager = (*repository)(nil)
)

// BuildSessionContextDataForUser queries the database for the memberships of a user and constructs a ContextData struct from the results.
func (r *repository) BuildSessionContextDataForUser(ctx context.Context, userID string) (*sessions.ContextData, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValue(platformkeys.UserIDKey, userID)
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)

	results, err := r.generatedQuerier.GetAccountUserMembershipsForUser(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user's memberships from database")
	}

	accountRolesMap := map[string]authorization.AccountRolePermissionsChecker{}
	defaultAccountID := ""
	for _, result := range results {
		accountRolesMap[result.BelongsToAccount] = authorization.NewAccountRolePermissionChecker(result.AccountRole)
		if result.DefaultAccount {
			defaultAccountID = result.BelongsToAccount
		}
	}

	user, err := r.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user from database")
	}

	sessionCtxData := &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:                   user.ID,
			Username:                 user.Username,
			EmailAddress:             user.EmailAddress,
			AccountStatus:            user.AccountStatus,
			AccountStatusExplanation: user.AccountStatusExplanation,
			ServicePermissions:       authorization.NewServiceRolePermissionChecker(user.ServiceRole),
		},
		AccountPermissions: accountRolesMap,
		ActiveAccountID:    defaultAccountID,
	}

	logger.Debug("fetched session context data for user")

	return sessionCtxData, nil
}

// GetDefaultAccountIDForUser retrieves the default account ID for a user.
func (r *repository) GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return "", platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)
	logger = logger.WithValue(platformkeys.UserIDKey, userID)

	id, err := r.generatedQuerier.GetDefaultAccountIDForUser(ctx, r.readDB, userID)
	if err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "fetching default account ID for user")
	}

	return id, nil
}

// markAccountAsUserDefault marks a given account as the user's default.
func (r *repository) markAccountAsUserDefault(ctx context.Context, querier database.SQLQueryExecutor, userID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValues(map[string]any{
		platformkeys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	if err = r.generatedQuerier.MarkAccountUserMembershipAsUserDefault(ctx, querier, &generated.MarkAccountUserMembershipAsUserDefaultParams{
		BelongsToUser:    userID,
		BelongsToAccount: accountID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "assigning user default account")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountUserMemberships,
		EventType:        audit.AuditLogEventTypeUpdated,
		BelongsToUser:    userID,
		Changes: map[string]*audit.ChangeLog{
			"default_account": {
				OldValue: "false",
				NewValue: "true",
			},
		},
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	logger.Debug("account marked as default")

	return nil
}

// MarkAccountAsUserDefault does a thing.
func (r *repository) MarkAccountAsUserDefault(ctx context.Context, userID, accountID string) error {
	return r.markAccountAsUserDefault(ctx, r.writeDB, userID, accountID)
}

// UserIsMemberOfAccount does a thing.
func (r *repository) UserIsMemberOfAccount(ctx context.Context, userID, accountID string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	result, err := r.generatedQuerier.UserIsAccountMember(ctx, r.readDB, &generated.UserIsAccountMemberParams{
		BelongsToAccount: accountID,
		BelongsToUser:    userID,
	})
	if err != nil {
		return false, observability.PrepareError(err, span, "performing user membership check query")
	}

	return result, nil
}

// ModifyUserPermissions does a thing.
func (r *repository) ModifyUserPermissions(ctx context.Context, accountID, userID string, input *identity.ModifyUserPermissionsInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := r.logger.WithValues(map[string]any{
		keys.AccountIDKey: accountID,
		platformkeys.UserIDKey:    userID,
		"new_roles":       input.NewRole,
	})

	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	memberships, err := r.generatedQuerier.GetAccountUserMembershipsForUser(ctx, tx, userID)
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "fetching user account memberships")
	}

	var existingRole string
	for _, membership := range memberships {
		if membership.BelongsToUser == userID && membership.BelongsToAccount == accountID {
			existingRole = membership.AccountRole
			break
		}
	}

	// modify the membership.
	if err = r.generatedQuerier.ModifyAccountUserPermissions(ctx, tx, &generated.ModifyAccountUserPermissionsParams{
		AccountRole:      input.NewRole,
		BelongsToAccount: accountID,
		BelongsToUser:    userID,
	}); err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "modifying user account permissions")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountUserMemberships,
		EventType:        audit.AuditLogEventTypeUpdated,
		BelongsToUser:    userID,
		Changes: map[string]*audit.ChangeLog{
			"account_role": {
				OldValue: existingRole,
				NewValue: input.NewRole,
			},
		},
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	logger.Info("user permissions modified")

	return nil
}

// TransferAccountOwnership does a thing.
func (r *repository) TransferAccountOwnership(ctx context.Context, accountID string, input *identity.AccountOwnershipTransferInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := r.logger.WithValues(map[string]any{
		keys.AccountIDKey: accountID,
		"current_owner":   input.CurrentOwner,
		"new_owner":       input.NewOwner,
	})

	tracing.AttachToSpan(span, platformkeys.UserIDKey, input.NewOwner)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	// create the membership.
	if err = r.generatedQuerier.TransferAccountOwnership(ctx, tx, &generated.TransferAccountOwnershipParams{
		NewOwner:  input.NewOwner,
		OldOwner:  input.CurrentOwner,
		AccountID: accountID,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "transferring account to new owner")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountUserMemberships,
		EventType:        audit.AuditLogEventTypeUpdated,
		BelongsToUser:    input.NewOwner,
		Changes: map[string]*audit.ChangeLog{
			"belongs_to_user": {
				OldValue: input.CurrentOwner,
				NewValue: input.NewOwner,
			},
		},
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	isMember, err := r.UserIsMemberOfAccount(ctx, input.NewOwner, accountID)
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "checking if user is member of account")
	}

	if !isMember {
		if err = r.addUserToAccount(ctx, tx, &identity.AccountUserMembershipDatabaseCreationInput{
			ID:          identifiers.New(),
			Reason:      "transferred ownership",
			UserID:      input.NewOwner,
			AccountID:   accountID,
			AccountRole: "account_admin",
		}); err != nil {
			r.RollbackTransaction(ctx, tx)
			return observability.PrepareAndLogError(err, logger, span, "adding user to account")
		}
		// audit log created above
	}

	if err = r.removeUserFromAccount(ctx, tx, input.CurrentOwner, accountID); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "removing user from account")
	}

	// audit log created above

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	logger.Info("account transferred to new owner")

	return nil
}

// addUserToAccount does a thing.
func (r *repository) addUserToAccount(ctx context.Context, querier database.SQLQueryExecutor, input *identity.AccountUserMembershipDatabaseCreationInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := r.logger.WithValues(map[string]any{
		platformkeys.UserIDKey:    input.UserID,
		keys.AccountIDKey: input.AccountID,
	})

	tracing.AttachToSpan(span, platformkeys.UserIDKey, input.UserID)
	tracing.AttachToSpan(span, keys.AccountIDKey, input.AccountID)

	// create the membership.
	if err := r.generatedQuerier.AddUserToAccount(ctx, querier, &generated.AddUserToAccountParams{
		ID:               input.ID,
		BelongsToUser:    input.UserID,
		BelongsToAccount: input.AccountID,
		AccountRole:      input.AccountRole,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "performing user account membership creation query")
	}

	if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, querier, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &input.AccountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountUserMemberships,
		EventType:        audit.AuditLogEventTypeCreated,
		BelongsToUser:    input.UserID,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	logger.Info("user added to account")

	return nil
}

// removeUserFromAccount removes a user's membership to an account.
func (r *repository) removeUserFromAccount(ctx context.Context, querier database.SQLQueryExecutorAndTransactionManager, userID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	logger := r.logger.WithValues(map[string]any{
		platformkeys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	// remove the membership.
	if err := r.generatedQuerier.RemoveUserFromAccount(ctx, querier, &generated.RemoveUserFromAccountParams{
		BelongsToAccount: accountID,
		BelongsToUser:    userID,
	}); err != nil {
		r.RollbackTransaction(ctx, querier)
		return observability.PrepareAndLogError(err, logger, span, "removing user from account")
	}

	if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, querier, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountUserMemberships,
		EventType:        audit.AuditLogEventTypeArchived,
		BelongsToUser:    userID,
	}); err != nil {
		r.RollbackTransaction(ctx, querier)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	remainingAccounts, err := r.getAccountsForUser(ctx, querier, userID, nil)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.RollbackTransaction(ctx, querier)
		return observability.PrepareError(err, span, "fetching remaining accounts")
	}

	if remainingAccounts == nil || len(remainingAccounts.Data) < 1 {
		logger.Debug("user has no remaining accounts, creating a new one")
		if _, err = r.createAccountForUser(ctx, querier, false, "", userID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating account for new user")
		}
		return nil
	}

	account := remainingAccounts.Data[0]
	logger.WithValue("new_default_account", account.ID).Info("setting new default account")
	if err = r.markAccountAsUserDefault(ctx, querier, userID, account.ID); err != nil {
		r.RollbackTransaction(ctx, querier)
		return observability.PrepareAndLogError(err, logger, span, "marking account as default")
	}

	logger.Info("marked account as default")

	return nil
}

// RemoveUserFromAccount removes a user's membership to an account.
func (r *repository) RemoveUserFromAccount(ctx context.Context, userID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	logger := r.logger.WithValues(map[string]any{
		platformkeys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, span, "beginning transaction")
	}

	if err = r.removeUserFromAccount(ctx, tx, userID, accountID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing user from account")
	}

	// audit log entry created above

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	return nil
}
