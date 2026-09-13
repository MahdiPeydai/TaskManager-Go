package validators

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValidationErrors(t *testing.T) {
	validate := validator.New()

	type request struct {
		Username string `validate:"required,min=3"`
		Email    string `validate:"required,email"`
	}

	t.Run("returns validation errors", func(t *testing.T) {
		req := request{}

		err := validate.Struct(req)
		require.Error(t, err)

		result := GetValidationErrors(err)

		require.Len(t, result, 2)

		assert.Equal(t, ValidationError{
			Property: "Username",
			Tag:      "required",
			Value:    "",
			Message:  "",
		}, result[0])

		assert.Equal(t, ValidationError{
			Property: "Email",
			Tag:      "required",
			Value:    "",
			Message:  "",
		}, result[1])
	})

	t.Run("returns nil for non-validation error", func(t *testing.T) {
		err := errors.New("some error")

		result := GetValidationErrors(err)

		assert.Nil(t, result)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		result := GetValidationErrors(nil)

		assert.Nil(t, result)
	})
}
