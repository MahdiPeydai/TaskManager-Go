package services

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTokenService() *TokenService {
	cfg := &config.Config{}

	cfg.Jwt.Secret = "test-secret"
	cfg.Jwt.AccessTokenExpireTime = 3600
	cfg.Jwt.RefreshTokenExpireTime = 86400

	logger := &mocks.MockLogger{}

	return GetTokenService(cfg, logger)
}

func TestTokenService_GenerateToken(t *testing.T) {
	service := testTokenService()

	firstName := "John"
	lastName := "Doe"
	mobile := "09123456789"
	email := "john@example.com"

	tokenData := Token{
		UserId:       123,
		FirstName:    &firstName,
		LastName:     &lastName,
		Username:     "john",
		MobileNumber: &mobile,
		Email:        &email,
		Roles:        []string{"user", "admin"},
	}

	before := time.Now().Unix()

	result, err := service.GenerateToken(tokenData)

	require.NoError(t, err)
	require.NotNil(t, result)

	after := time.Now().Unix()

	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	// Access token has "Bearer " prefix.
	assert.True(
		t,
		strings.HasPrefix(
			result.AccessToken,
			constants.AuthorizationHeaderPrefix+" ",
		),
	)

	// Check expiration timestamps.
	assert.GreaterOrEqual(t, result.AccessTokenExpireTime, before+3599)
	assert.LessOrEqual(t, result.AccessTokenExpireTime, after+3601)

	assert.GreaterOrEqual(t, result.RefreshTokenExpireTime, before+86399)
	assert.LessOrEqual(t, result.RefreshTokenExpireTime, after+86401)

	// Verify the access token.
	accessToken := strings.TrimPrefix(
		result.AccessToken,
		constants.AuthorizationHeaderPrefix+" ",
	)

	verifiedAccessToken, err := service.VerifyToken(accessToken)

	require.NoError(t, err)
	require.NotNil(t, verifiedAccessToken)
	assert.True(t, verifiedAccessToken.Valid)

	// Verify the refresh token.
	verifiedRefreshToken, err := service.VerifyToken(result.RefreshToken)

	require.NoError(t, err)
	require.NotNil(t, verifiedRefreshToken)
	assert.True(t, verifiedRefreshToken.Valid)
}

func TestTokenService_GenerateToken_Claims(t *testing.T) {
	service := testTokenService()

	firstName := "John"
	lastName := "Doe"

	tokenData := Token{
		UserId:    123,
		FirstName: &firstName,
		LastName:  &lastName,
		Username:  "john",
		Roles:     []string{"user", "admin"},
	}

	result, err := service.GenerateToken(tokenData)

	require.NoError(t, err)

	accessToken := strings.TrimPrefix(
		result.AccessToken,
		constants.AuthorizationHeaderPrefix+" ",
	)

	claims, err := service.GetClaims(accessToken)

	require.NoError(t, err)

	assert.Equal(t, float64(123), claims[constants.UserIdKey])
	assert.Equal(t, "John", claims[constants.FirstNameKey])
	assert.Equal(t, "Doe", claims[constants.LastNameKey])
	assert.Equal(t, "john", claims[constants.UsernameKey])
	assert.Equal(t, []interface{}{"user", "admin"}, claims[constants.RolesKey])
	assert.Equal(t, float64(result.AccessTokenExpireTime), claims[constants.ExpireTimeKey])
	assert.Equal(t, constants.AccessToken, claims[constants.TokenTypeKey])
}

func TestTokenService_GenerateToken_RefreshTokenClaims(t *testing.T) {
	service := testTokenService()

	tokenData := Token{
		UserId:   123,
		Username: "john",
		Roles:    []string{"user"},
	}

	result, err := service.GenerateToken(tokenData)

	require.NoError(t, err)

	claims, err := service.GetClaims(result.RefreshToken)

	require.NoError(t, err)

	assert.Equal(t, float64(123), claims[constants.UserIdKey])
	assert.Equal(t, constants.RefreshToken, claims[constants.TokenTypeKey])
}

func TestTokenService_VerifyToken(t *testing.T) {
	service := testTokenService()

	validTokenData := Token{
		UserId:   123,
		Username: "john",
		Roles:    []string{"user"},
	}

	generated, err := service.GenerateToken(validTokenData)
	require.NoError(t, err)

	validAccessToken := strings.TrimPrefix(
		generated.AccessToken,
		constants.AuthorizationHeaderPrefix+" ",
	)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token",
			token:       validAccessToken,
			expectError: false,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name: "tampered token",
			token: func() string {
				parts := strings.Split(validAccessToken, ".")
				return parts[0] + "." + parts[1] + ".tampered"
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.VerifyToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.Valid)
		})
	}
}

func TestTokenService_VerifyToken_RejectsNonHMAC(t *testing.T) {
	service := testTokenService()

	token := jwt.NewWithClaims(
		jwt.SigningMethodNone,
		jwt.MapClaims{
			"user_id": 123,
		},
	)

	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	result, err := service.VerifyToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestTokenService_VerifyToken_WrongSecret(t *testing.T) {
	service := testTokenService()

	tokenData := Token{
		UserId:   123,
		Username: "john",
	}

	generated, err := service.GenerateToken(tokenData)
	require.NoError(t, err)

	accessToken := strings.TrimPrefix(
		generated.AccessToken,
		constants.AuthorizationHeaderPrefix+" ",
	)

	otherCfg := &config.Config{}
	otherCfg.Jwt.Secret = "different-secret"
	otherCfg.Jwt.AccessTokenExpireTime = 3600
	otherCfg.Jwt.RefreshTokenExpireTime = 86400

	otherService := GetTokenService(otherCfg, &mocks.MockLogger{})

	result, err := otherService.VerifyToken(accessToken)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestTokenService_GetClaims(t *testing.T) {
	service := testTokenService()

	tokenData := Token{
		UserId:   123,
		Username: "john",
		Roles:    []string{"user"},
	}

	generated, err := service.GenerateToken(tokenData)
	require.NoError(t, err)

	accessToken := strings.TrimPrefix(
		generated.AccessToken,
		constants.AuthorizationHeaderPrefix+" ",
	)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token",
			token:       accessToken,
			expectError: false,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.GetClaims(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)

			assert.Equal(t, float64(123), claims[constants.UserIdKey])
			assert.Equal(t, "john", claims[constants.UsernameKey])
		})
	}
}
