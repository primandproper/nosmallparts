package helpers

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func FetchErrorString(err error, key string) string {
	var validErr validation.Errors
	if errors.As(err, &validErr) {
		if validationErr := validErr[key]; validationErr != nil {
			var validationLibError validation.ErrorObject
			if errors.As(validationErr, &validationLibError) {
				return validationLibError.Error()
			}
		}
	}

	return ""
}
