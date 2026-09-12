package validators

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Property string
	Tag      string
	Value    string
	Message  string
}

func GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Property = err.Field()
			element.Tag = err.Tag()
			element.Value = err.Param()
			validationErrors = append(validationErrors, element)
		}
		return validationErrors
	}
	return nil
}
