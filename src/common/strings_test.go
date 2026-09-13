package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "Password123!"

	hashedPassword, err := HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	assert.True(t, ComparePasswords(hashedPassword, password))
}

func TestHashPassword_GeneratesDifferentHashes(t *testing.T) {
	password := "Password123!"

	firstHash, err := HashPassword(password)
	require.NoError(t, err)

	secondHash, err := HashPassword(password)
	require.NoError(t, err)

	assert.NotEqual(t, firstHash, secondHash)

	assert.True(t, ComparePasswords(firstHash, password))
	assert.True(t, ComparePasswords(secondHash, password))
}

func TestComparePasswords(t *testing.T) {
	hashedPassword, err := HashPassword("Password123!")
	require.NoError(t, err)

	tests := []struct {
		name     string
		hashedPw string
		plainPw  string
		expected bool
	}{
		{
			name:     "correct password",
			hashedPw: hashedPassword,
			plainPw:  "Password123!",
			expected: true,
		},
		{
			name:     "incorrect password",
			hashedPw: hashedPassword,
			plainPw:  "WrongPassword123!",
			expected: false,
		},
		{
			name:     "invalid hash",
			hashedPw: "invalid-hash",
			plainPw:  "Password123!",
			expected: false,
		},
		{
			name:     "empty password",
			hashedPw: hashedPassword,
			plainPw:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ComparePasswords(tt.hashedPw, tt.plainPw))
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name             string
		password         string
		minLength        int
		maxLength        int
		includeChars     bool
		includeDigits    bool
		includeLowercase bool
		includeUppercase bool
		expected         bool
	}{
		{
			name:             "valid password",
			password:         "Password123",
			minLength:        8,
			maxLength:        20,
			includeChars:     true,
			includeDigits:    true,
			includeLowercase: true,
			includeUppercase: true,
			expected:         true,
		},
		{
			name:             "password too short",
			password:         "Pass1",
			minLength:        8,
			maxLength:        20,
			includeChars:     true,
			includeDigits:    true,
			includeLowercase: true,
			includeUppercase: true,
			expected:         false,
		},
		{
			name:             "password too long",
			password:         "Password123456789012345",
			minLength:        8,
			maxLength:        10,
			includeChars:     true,
			includeDigits:    true,
			includeLowercase: true,
			includeUppercase: true,
			expected:         false,
		},
		{
			name:             "letters required but missing",
			password:         "12345678",
			minLength:        8,
			maxLength:        20,
			includeChars:     true,
			includeDigits:    false,
			includeLowercase: false,
			includeUppercase: false,
			expected:         false,
		},
		{
			name:             "digits required but missing",
			password:         "Password",
			minLength:        8,
			maxLength:        20,
			includeChars:     true,
			includeDigits:    true,
			includeLowercase: false,
			includeUppercase: false,
			expected:         false,
		},
		{
			name:             "lowercase required but missing",
			password:         "PASSWORD123",
			minLength:        8,
			maxLength:        20,
			includeChars:     true,
			includeDigits:    true,
			includeLowercase: true,
			includeUppercase: true,
			expected:         false,
		},
		{
			name:             "uppercase required but missing",
			password:         "password123",
			minLength:        8,
			maxLength:        20,
			includeChars:     true,
			includeDigits:    true,
			includeLowercase: true,
			includeUppercase: true,
			expected:         false,
		},
		{
			name:             "only length validation",
			password:         "12345678",
			minLength:        8,
			maxLength:        20,
			includeChars:     false,
			includeDigits:    false,
			includeLowercase: false,
			includeUppercase: false,
			expected:         true,
		},
		{
			name:             "all requirements disabled",
			password:         "abcdefgh",
			minLength:        8,
			maxLength:        20,
			includeChars:     false,
			includeDigits:    false,
			includeLowercase: false,
			includeUppercase: false,
			expected:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(
				tt.password,
				tt.minLength,
				tt.maxLength,
				tt.includeChars,
				tt.includeDigits,
				tt.includeLowercase,
				tt.includeUppercase,
			)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "contains uppercase",
			input:    "helloWorld",
			expected: true,
		},
		{
			name:     "only lowercase",
			input:    "helloworld",
			expected: false,
		},
		{
			name:     "only digits",
			input:    "123456",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasUpper(tt.input))
		})
	}
}

func TestHasLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "contains lowercase",
			input:    "HelloWorld",
			expected: true,
		},
		{
			name:     "only uppercase",
			input:    "HELLOWORLD",
			expected: false,
		},
		{
			name:     "only digits",
			input:    "123456",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasLower(tt.input))
		})
	}
}

func TestHasLetter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "contains letter",
			input:    "123a456",
			expected: true,
		},
		{
			name:     "only digits",
			input:    "123456",
			expected: false,
		},
		{
			name:     "only special characters",
			input:    "!@#$%",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasLetter(tt.input))
		})
	}
}

func TestHasDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "contains digit",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "only letters",
			input:    "abcdef",
			expected: false,
		},
		{
			name:     "only special characters",
			input:    "!@#$%",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasDigits(tt.input))
		})
	}
}
