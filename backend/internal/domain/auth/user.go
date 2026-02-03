package auth

import (
	"context"
	"errors"
	"math"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	validTOTPTokenLength = 6
)

var (
	totpTokenLengthRule = validation.Length(validTOTPTokenLength, validTOTPTokenLength)

	errNewPasswordSameAsOld = errors.New("new password cannot be the same as the old password")
)

type (
	// UserLoginInput represents the payload used to log in a User.
	UserLoginInput struct {
		_ struct{} `json:"-"`

		Username  string `json:"username"`
		Password  string `json:"password"`
		TOTPToken string `json:"totpToken"`
	}

	// PasswordUpdateInput represents input a User would provide when updating their passwords.
	PasswordUpdateInput struct {
		_ struct{} `json:"-"`

		NewPassword     string `json:"newPassword"`
		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// UsernameUpdateInput represents input a User would provide when updating their username.
	UsernameUpdateInput struct {
		_ struct{} `json:"-"`

		NewUsername     string `json:"newUsername"`
		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// UserEmailAddressUpdateInput represents input a User would provide when updating their email address.
	UserEmailAddressUpdateInput struct {
		_ struct{} `json:"-"`

		NewEmailAddress string `json:"newEmailAddress"`
		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// TOTPSecretRefreshInput represents input a User would provide when updating their 2FA secret.
	TOTPSecretRefreshInput struct {
		_ struct{} `json:"-"`

		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// TOTPSecretVerificationInput represents input a User would provide when validating their 2FA secret.
	TOTPSecretVerificationInput struct {
		_ struct{} `json:"-"`

		TOTPToken string `json:"totpToken"`
		UserID    string `json:"userID"`
	}

	// TOTPSecretRefreshResponse represents the response we provide to a User when updating their 2FA secret.
	TOTPSecretRefreshResponse struct {
		_ struct{} `json:"-"`

		TwoFactorQRCode string `json:"qrCode"`
		TwoFactorSecret string `json:"twoFactorSecret"`
	}

	// EmailAddressVerificationRequestInput represents the request a User provides when verifying their email address.
	EmailAddressVerificationRequestInput struct {
		_ struct{} `json:"-"`

		Token string `json:"emailVerificationToken"`
	}

	// PasswordResetResponse is returned when a user updates their password.
	PasswordResetResponse struct {
		Successful bool `json:"successful,omitempty"`
	}

	// UserAccountStatusUpdateInput represents what an admin User could provide as input for changing statuses.
	UserAccountStatusUpdateInput struct {
		_ struct{} `json:"-"`

		NewStatus    string `json:"newStatus"`
		Reason       string `json:"reason"`
		TargetUserID string `json:"targetUserID"`
	}
)

var _ validation.ValidatableWithContext = (*TOTPSecretVerificationInput)(nil)

// ValidateWithContext ensures our provided TOTPSecretVerificationInput meets expectations.
func (i *TOTPSecretVerificationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.UserID, validation.Required),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

// ValidateWithContext ensures our provided UserLoginInput meets expectations.
func (i *UserLoginInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Username, validation.Required, validation.Length(4, math.MaxInt8)),
		validation.Field(&i.Password, validation.Required, validation.Length(8, math.MaxInt8)),
		validation.Field(&i.TOTPToken, is.Digit, validation.RuneLength(6, 6)),
	)
}

// ValidateWithContext ensures our provided PasswordUpdateInput meets expectations.
func (i *PasswordUpdateInput) ValidateWithContext(ctx context.Context, minPasswordLength uint8) error {
	if i.CurrentPassword == i.NewPassword {
		return errNewPasswordSameAsOld
	}

	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.CurrentPassword, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
		validation.Field(&i.NewPassword, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
	)
}

var _ validation.ValidatableWithContext = (*TOTPSecretRefreshInput)(nil)

// ValidateWithContext ensures our provided TOTPSecretRefreshInput meets expectations.
func (i *TOTPSecretRefreshInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.CurrentPassword, validation.Required),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

var _ validation.ValidatableWithContext = (*EmailAddressVerificationRequestInput)(nil)

// ValidateWithContext ensures our provided EmailAddressVerificationRequestInput meets expectations.
func (i *EmailAddressVerificationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Token, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*UsernameUpdateInput)(nil)

// ValidateWithContext ensures our provided UsernameUpdateInput meets expectations.
func (i *UsernameUpdateInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.NewUsername, validation.Required),
		validation.Field(&i.CurrentPassword, validation.Required),
		validation.Field(&i.TOTPToken, validation.When(i.TOTPToken != "", totpTokenLengthRule)),
	)
}

var _ validation.ValidatableWithContext = (*UserEmailAddressUpdateInput)(nil)

// ValidateWithContext ensures our provided UserEmailAddressUpdateInput meets expectations.
func (i *UserEmailAddressUpdateInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.NewEmailAddress, validation.Required),
		validation.Field(&i.CurrentPassword, validation.Required),
		validation.Field(&i.TOTPToken, validation.When(i.TOTPToken != "", totpTokenLengthRule)),
	)
}
