package validators

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPasswordValidator(t *testing.T) {
	cfg := &config.Config{}
	cfg.Password.MinLength = 8
	cfg.Password.MaxLength = 20
	cfg.Password.IncludeChars = true
	cfg.Password.IncludeDigits = true
	cfg.Password.IncludeLowercase = true
	cfg.Password.IncludeUppercase = true

	validate := validator.New()
	require.NoError(t, validate.RegisterValidation(
		"user_password",
		UserPasswordValidator(cfg),
	))

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "valid password",
			password: "Password123",
			expected: true,
		},
		{
			name:     "too short",
			password: "Pass1",
			expected: false,
		},
		{
			name:     "missing digit",
			password: "Password",
			expected: false,
		},
		{
			name:     "missing lowercase",
			password: "PASSWORD123",
			expected: false,
		},
		{
			name:     "missing uppercase",
			password: "password123",
			expected: false,
		},
		{
			name:     "missing letter",
			password: "12345678",
			expected: false,
		},
	}

	type request struct {
		Password string `validate:"user_password"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(request{
				Password: tt.password,
			})

			assert.Equal(t, tt.expected, err == nil)
		})
	}
}
