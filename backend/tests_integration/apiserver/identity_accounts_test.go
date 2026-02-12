package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	identityconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"
	identitygrpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultNumberOfAccountsAssociatedWithUsers = 1
)

func TestAccounts_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleAccount := fakes.BuildFakeAccountCreationRequestInput()
		exampleAccountInput := identitygrpcconverters.ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(exampleAccount)

		createdAccount, err := testClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{Input: exampleAccountInput})
		assert.NoError(t, err)
		assert.NotNil(t, createdAccount)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		testClient := buildUnauthenticatedHTTPClientForTest(t)

		exampleAccount := fakes.BuildFakeAccountCreationRequestInput()
		exampleAccountInput := identitygrpcconverters.ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(exampleAccount)

		createdAccount, err := testClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{Input: exampleAccountInput})
		assert.Error(t, err)
		assert.Nil(t, createdAccount)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleAccount := fakes.BuildFakeAccountCreationRequestInput()
		exampleAccountInput := identitygrpcconverters.ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(exampleAccount)
		// not allowed
		exampleAccountInput.Name = ""

		createdAccount, err := testClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{Input: exampleAccountInput})
		assert.Error(t, err)
		assert.Nil(t, createdAccount)
	})
}

func TestAccounts_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		var createdAccounts []*identitysvc.Account
		for range 5 {
			exampleAccount := fakes.BuildFakeAccountCreationRequestInput()
			exampleAccountInput := identitygrpcconverters.ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(exampleAccount)

			createdAccount, err := testClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{Input: exampleAccountInput})
			require.NoError(t, err)
			require.NotNil(t, createdAccount)

			createdAccounts = append(createdAccounts, createdAccount.Created)
		}

		accounts, err := testClient.GetAccounts(ctx, &identitysvc.GetAccountsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, accounts)
		assert.Equal(t, len(accounts.Results), len(createdAccounts)+defaultNumberOfAccountsAssociatedWithUsers)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		_, _ = createUserAndClientForTest(t)
		testClient := buildUnauthenticatedHTTPClientForTest(t)

		_, err := testClient.GetAccounts(ctx, &identitysvc.GetAccountsRequest{})
		assert.Error(t, err)
	})
}

func TestAccounts_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleAccount := fakes.BuildFakeAccountCreationRequestInput()
		exampleAccountInput := identitygrpcconverters.ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(exampleAccount)

		createdAccount, err := testClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{Input: exampleAccountInput})
		require.NoError(t, err)
		require.NotNil(t, createdAccount)

		retrievedAccount, err := testClient.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: createdAccount.Created.Id})
		assert.NoError(t, err)
		assert.NotNil(t, createdAccount)

		converted := identitygrpcconverters.ConvertGRPCAccountToAccount(retrievedAccount.Result)

		assertRoughEquality(t, identitygrpcconverters.ConvertGRPCAccountToAccount(createdAccount.Created), converted, append(defaultIgnoredFields(), "Members")...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		_, testClient := createUserAndClientForTest(t)

		// fetch the account
		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		testClient = buildUnauthenticatedHTTPClientForTest(t)

		_, err = testClient.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: account.Result.Id})
		assert.Error(t, err)
	})

	T.Run("for nonexistent account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrievedAccount, err := testClient.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrievedAccount)
	})
}

func TestAccounts_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleAccount := fakes.BuildFakeAccountCreationRequestInput()
		exampleAccountInput := identitygrpcconverters.ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(exampleAccount)

		createdAccount, err := testClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{Input: exampleAccountInput})
		require.NoError(t, err)
		require.NotNil(t, createdAccount)

		converted := identitygrpcconverters.ConvertGRPCAccountToAccount(createdAccount.Created)
		converted.Name = "Updated name"

		_, err = testClient.SetDefaultAccount(ctx, &identitysvc.SetDefaultAccountRequest{AccountId: converted.ID})
		require.NoError(t, err)

		updateInput := identityconverters.ConvertAccountToAccountUpdateRequestInput(converted)
		_, err = testClient.UpdateAccount(ctx, &identitysvc.UpdateAccountRequest{
			AccountId: converted.ID,
			Input:     identitygrpcconverters.ConvertAccountUpdateRequestInputToGRPCAccountUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updatedClient, err := testClient.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: converted.ID})
		assert.NoError(t, err)
		assert.NotNil(t, updatedClient)
		assert.Equal(t, converted.Name, updatedClient.Result.Name)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		_, testClient := createUserAndClientForTest(t)

		// fetch the account
		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		testClient = buildUnauthenticatedHTTPClientForTest(t)

		_, err = testClient.UpdateAccount(ctx, &identitysvc.UpdateAccountRequest{AccountId: account.Result.Id})
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		/*
			there's no way to provide invalid input to this method, but
			I want to make it explicit that tests should be written the moment that changes
		*/
	})

	T.Run("for nonexistent account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		converted := fakes.BuildFakeAccount()
		updateInput := identityconverters.ConvertAccountToAccountUpdateRequestInput(converted)
		_, err := testClient.UpdateAccount(ctx, &identitysvc.UpdateAccountRequest{
			AccountId: nonexistentID,
			Input:     identitygrpcconverters.ConvertAccountUpdateRequestInputToGRPCAccountUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updatedClient, err := testClient.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: converted.ID})
		assert.Error(t, err)
		assert.Nil(t, updatedClient)
	})
}

func TestAccounts_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleAccount := fakes.BuildFakeAccountCreationRequestInput()
		exampleAccountInput := identitygrpcconverters.ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(exampleAccount)

		createdAccount, err := testClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{Input: exampleAccountInput})
		require.NoError(t, err)
		require.NotNil(t, createdAccount)

		_, err = testClient.ArchiveAccount(ctx, &identitysvc.ArchiveAccountRequest{AccountId: createdAccount.Created.Id})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		_, testClient := createUserAndClientForTest(t)

		// fetch the account
		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		testClient = buildUnauthenticatedHTTPClientForTest(t)

		_, err = testClient.ArchiveAccount(ctx, &identitysvc.ArchiveAccountRequest{AccountId: account.Result.Id})
		assert.Error(t, err)
	})

	T.Run("for nonexistent account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveAccount(ctx, &identitysvc.ArchiveAccountRequest{AccountId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestAccounts_Inviting(T *testing.T) {
	T.Parallel()

	T.Run("invite user via their email address", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account ID to send invites for
		_, testClient := createUserAndClientForTest(t)
		accountRes, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		accountID := accountRes.Result.Id

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// create a user to invite
		inviteeEmailAddress := fmt.Sprintf("some_fake_email%d@testing.com", time.Now().UnixMicro())
		input := &identity.UserRegistrationInput{
			Birthday:              pointer.To(time.Now()),
			EmailAddress:          inviteeEmailAddress,
			FirstName:             fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AccountName:           fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			LastName:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Password:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Username:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AcceptedPrivacyPolicy: true,
			AcceptedTOS:           true,
		}
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// create the invitation for the user
		invitation, err := testClient.CreateAccountInvitation(ctx, &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				Note:    t.Name(),
				ToName:  t.Name(),
				ToEmail: inviteeEmailAddress,
			},
		})
		require.NoError(t, err)

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.GetSentAccountInvitations(ctx, &identitysvc.GetSentAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.Results)

		// verify the invitee can see the invitation as received
		invitations, err := inviteeClient.GetReceivedAccountInvitations(ctx, &identitysvc.GetReceivedAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.NotEmpty(t, invitations.Results)

		// accept the invitation
		_, err = inviteeClient.AcceptAccountInvitation(ctx, &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: invitation.Created.Id,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: invitation.Created.Token,
				Note:  t.Name(),
			},
		})
		require.NoError(t, err)

		// the invited user needs a new token that indicates they're a member of this account
		inviteeClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, invitee))
		require.NoError(t, err)

		// verify that we don't have any sent invitations because they've all been accepted
		sentInvitations, err = testClient.GetSentAccountInvitations(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.Results)

		// verify that the invited user can see the account in their accounts list
		accounts, err := inviteeClient.GetAccounts(ctx, &identitysvc.GetAccountsRequest{})
		require.NoError(t, err)
		require.NotNil(t, accounts)
		assert.Len(t, accounts.Results, 2)

		var found bool
		for _, account := range accounts.Results {
			if !found {
				found = account.Id == accountID
			}
		}
		require.True(t, found)

		// change to the new account
		_, err = inviteeClient.SetDefaultAccount(ctx, &identitysvc.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		// validate we can see the webhook created before our user existed
		webhook, err := inviteeClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.ID})
		require.NoError(t, err)
		require.NotNil(t, webhook)
	})

	T.Run("invite user via token and invite MealPlanTaskID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account MealPlanTaskID to send invites for
		_, testClient := createUserAndClientForTest(t)
		accountRes, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		accountID := accountRes.Result.Id

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// create the invitation for the user
		invitation, err := testClient.CreateAccountInvitation(ctx, &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				Note:   t.Name(),
				ToName: t.Name(),
			},
		})
		require.NoError(t, err)

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.GetSentAccountInvitations(ctx, &identitysvc.GetSentAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.Results)

		// create a user to invite
		inviteeEmailAddress := fmt.Sprintf("some_fake_email%d@testing.com", time.Now().UnixMicro())
		input := &identity.UserRegistrationInput{
			Birthday:              pointer.To(time.Now()),
			EmailAddress:          inviteeEmailAddress,
			FirstName:             fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AccountName:           fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			LastName:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Password:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Username:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			InvitationID:          invitation.Created.Id,
			InvitationToken:       invitation.Created.Token,
			AcceptedPrivacyPolicy: true,
			AcceptedTOS:           true,
		}
		_, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// verify the invitee can see the invitation as received
		invitations, err := inviteeClient.GetReceivedAccountInvitations(ctx, &identitysvc.GetReceivedAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.Empty(t, invitations.Results)

		// verify that we don't have any sent invitations because they've all been accepted
		sentInvitations, err = testClient.GetSentAccountInvitations(ctx, &identitysvc.GetSentAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.Results)

		// verify that the invited user can see the account in their accounts list
		accounts, err := inviteeClient.GetAccounts(ctx, &identitysvc.GetAccountsRequest{})
		require.NoError(t, err)
		require.NotNil(t, accounts)
		assert.Len(t, accounts.Results, 2)

		var found bool
		for _, account := range accounts.Results {
			if !found {
				found = account.Id == accountID
			}
		}
		require.True(t, found)

		// change to the new account
		_, err = inviteeClient.SetDefaultAccount(ctx, &identitysvc.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		// validate we can see the webhook created before our user existed
		webhook, err := inviteeClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.ID})
		require.NoError(t, err)
		require.NotNil(t, webhook)
	})

	T.Run("invites can be canceled before acceptance", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account MealPlanTaskID to send invites for
		_, testClient := createUserAndClientForTest(t)

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// create a user to invite
		inviteeEmailAddress := fmt.Sprintf("some_fake_email%d@testing.com", time.Now().UnixMicro())
		input := &identity.UserRegistrationInput{
			Birthday:              pointer.To(time.Now()),
			EmailAddress:          inviteeEmailAddress,
			FirstName:             fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AccountName:           fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			LastName:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Password:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Username:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AcceptedPrivacyPolicy: true,
			AcceptedTOS:           true,
		}
		_, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// create the invitation for the user
		invitation, err := testClient.CreateAccountInvitation(ctx, &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				Note:    t.Name(),
				ToName:  t.Name(),
				ToEmail: inviteeEmailAddress,
			},
		})
		require.NoError(t, err)

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.GetSentAccountInvitations(ctx, &identitysvc.GetSentAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.Results)

		// verify the invitee can see the invitation as received
		invitations, err := inviteeClient.GetReceivedAccountInvitations(ctx, &identitysvc.GetReceivedAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.NotEmpty(t, invitations.Results)

		_, err = testClient.CancelAccountInvitation(ctx, &identitysvc.CancelAccountInvitationRequest{
			AccountInvitationId: invitation.Created.Id,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: invitation.Created.Token,
				Note:  t.Name(),
			},
		})
		require.NoError(t, err)

		// accept the invitation
		_, err = inviteeClient.AcceptAccountInvitation(ctx, &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: invitation.Created.Id,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: invitation.Created.Token,
				Note:  t.Name(),
			},
		})
		assert.Error(t, err)

		// verify that we don't have any sent invitations because they've all been accepted
		sentInvitations, err = testClient.GetSentAccountInvitations(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.Results)

		// validate we can see the webhook created before our user existed
		webhook, err := inviteeClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.ID})
		assert.Error(t, err)
		assert.Nil(t, webhook)
	})

	T.Run("invites can be rejected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account MealPlanTaskID to send invites for
		_, testClient := createUserAndClientForTest(t)

		// create a user to invite
		inviteeEmailAddress := fmt.Sprintf("some_fake_email%d@testing.com", time.Now().UnixMicro())
		input := &identity.UserRegistrationInput{
			Birthday:              pointer.To(time.Now()),
			EmailAddress:          inviteeEmailAddress,
			FirstName:             fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AccountName:           fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			LastName:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Password:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Username:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AcceptedPrivacyPolicy: true,
			AcceptedTOS:           true,
		}
		_, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// create the invitation for the user
		invitation, err := testClient.CreateAccountInvitation(ctx, &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				Note:    t.Name(),
				ToName:  t.Name(),
				ToEmail: inviteeEmailAddress,
			},
		})
		require.NoError(t, err)

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.GetSentAccountInvitations(ctx, &identitysvc.GetSentAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.Results)

		// verify the invitee can see the invitation as received
		invitations, err := inviteeClient.GetReceivedAccountInvitations(ctx, &identitysvc.GetReceivedAccountInvitationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.NotEmpty(t, invitations.Results)

		// accept the invitation
		_, err = inviteeClient.RejectAccountInvitation(ctx, &identitysvc.RejectAccountInvitationRequest{
			AccountInvitationId: invitation.Created.Id,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: invitation.Created.Token,
				Note:  t.Name(),
			},
		})
		require.NoError(t, err)

		// verify that we don't have any sent invitations because they've all been accepted
		sentInvitations, err = testClient.GetSentAccountInvitations(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.Results)
	})
}

func TestAccounts_OwnershipTransfer(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account MealPlanTaskID to send invites for
		ogUser, testClient := createUserAndClientForTest(t)
		accountRes, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		accountID := accountRes.Result.Id

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// create a user to invite
		inviteeEmailAddress := fmt.Sprintf("some_fake_email%d@testing.com", time.Now().UnixMicro())
		input := &identity.UserRegistrationInput{
			Birthday:              pointer.To(time.Now()),
			EmailAddress:          inviteeEmailAddress,
			FirstName:             fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AccountName:           fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			LastName:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Password:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Username:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AcceptedPrivacyPolicy: true,
			AcceptedTOS:           true,
		}
		invitee, _ := createUserAndClientForTestWithRegistrationInput(t, input)

		// create the invitation for the user
		_, err = testClient.TransferAccountOwnership(ctx, &identitysvc.TransferAccountOwnershipRequest{
			AccountId: accountID,
			Input: &identitysvc.AccountOwnershipTransferInput{
				Reason:       t.Name(),
				CurrentOwner: ogUser.ID,
				NewOwner:     invitee.ID,
			},
		})
		require.NoError(t, err)

		// the invited user needs a new token that indicates they're a member of this account
		inviteeClient, err := buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, invitee))
		require.NoError(t, err)

		// change to the new account
		_, err = inviteeClient.SetDefaultAccount(ctx, &identitysvc.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		// validate we can see the webhook created before our user existed
		webhook, err := inviteeClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.ID})
		require.NoError(t, err)
		require.NotNil(t, webhook)
	})
}

func TestAccounts_UsersHaveBackupAccountCreatedForThemWhenRemovedFromLastAccount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account MealPlanTaskID to send invites for
		testUser, testClient := createUserAndClientForTest(t)
		accountRes, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		accountID := accountRes.Result.Id

		inviteeEmailAddress := fmt.Sprintf("some_fake_email%d@testing.com", time.Now().UnixMicro())
		inviteRes, err := testClient.CreateAccountInvitation(ctx, &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				Note:    t.Name(),
				ToEmail: inviteeEmailAddress,
				ToName:  fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, inviteRes)

		// create a user to invite
		input := &identity.UserRegistrationInput{
			Birthday:              pointer.To(time.Now()),
			EmailAddress:          inviteeEmailAddress,
			FirstName:             fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AccountName:           fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			LastName:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Password:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Username:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			InvitationID:          inviteRes.Created.Id,
			InvitationToken:       inviteRes.Created.Token,
			AcceptedPrivacyPolicy: true,
			AcceptedTOS:           true,
		}
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		inviteeAccountsRes, err := inviteeClient.GetAccounts(ctx, &identitysvc.GetAccountsRequest{})
		require.NoError(t, err)
		require.Len(t, inviteeAccountsRes.Results, 2)

		_, err = testClient.UpdateAccountMemberPermissions(ctx, &identitysvc.UpdateAccountMemberPermissionsRequest{
			UserId: invitee.ID,
			Input: &identitysvc.ModifyUserPermissionsInput{
				Reason:  t.Name(),
				NewRole: "account_admin",
			},
		})
		require.NoError(t, err)

		///////

		var (
			found          bool
			otherAccountID string
		)

		for _, account := range inviteeAccountsRes.Results {
			if account.Id == accountID {
				if !found {
					found = true
				}
			} else {
				otherAccountID = account.Id
			}
		}

		require.NotEmpty(t, otherAccountID)
		require.True(t, found)

		_, err = inviteeClient.ArchiveUserMembership(ctx, &identitysvc.ArchiveUserMembershipRequest{
			AccountId: accountID,
			UserId:    testUser.ID,
		})
		require.NoError(t, err)

		account, err := inviteeClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)
		assert.NotEqual(t, account, accountID)

		require.True(t, found)
	})
}
