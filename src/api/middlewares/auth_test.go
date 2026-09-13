package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func setupGin() (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()
	return r, w
}

// -----------------------------------------------------------------------------
// Authentication Middleware Tests
// -----------------------------------------------------------------------------

func TestAuthentication_MissingAuthorizationHeader(t *testing.T) {
	r, w := setupGin()
	cfg := &config.Config{} // Pass your app configuration mock/instance

	r.Use(Authentication(cfg, &mocks.MockLogger{}))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthentication_InvalidTokenFormat_MissingBearerPrefix(t *testing.T) {
	r, w := setupGin()
	cfg := &config.Config{}

	r.Use(Authentication(cfg, &mocks.MockLogger{}))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", http.NoBody)
	req.Header.Set(constants.AuthorizationHeaderKey, "InvalidFormatToken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthentication_InvalidTokenSchema_WrongPrefix(t *testing.T) {
	r, w := setupGin()
	cfg := &config.Config{}

	r.Use(Authentication(cfg, &mocks.MockLogger{}))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", http.NoBody)
	req.Header.Set(constants.AuthorizationHeaderKey, "Basic my-token-string")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// -----------------------------------------------------------------------------
// Authorization Middleware Tests
// -----------------------------------------------------------------------------

func TestAuthorization_MissingUserIdInContext(t *testing.T) {
	r, w := setupGin()

	r.Use(Authorization([]string{"admin"}))
	r.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorization_MissingRolesInContext(t *testing.T) {
	r, w := setupGin()

	r.Use(func(c *gin.Context) {
		// Set UserId but omit RolesKey
		c.Set(constants.UserIdKey, 1)
		c.Next()
	})
	r.Use(Authorization([]string{"admin"}))
	r.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorization_NilRolesInContext(t *testing.T) {
	r, w := setupGin()

	r.Use(func(c *gin.Context) {
		c.Set(constants.UserIdKey, 1)
		c.Set(constants.RolesKey, nil)
		c.Next()
	})
	r.Use(Authorization([]string{"admin"}))
	r.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorization_UserLacksRequiredRole(t *testing.T) {
	r, w := setupGin()

	r.Use(func(c *gin.Context) {
		c.Set(constants.UserIdKey, 1)
		c.Set(constants.RolesKey, []interface{}{"user", "guest"})
		c.Next()
	})
	r.Use(Authorization([]string{"admin", "manager"}))
	r.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorization_UserHasMatchingRole(t *testing.T) {
	r, w := setupGin()

	r.Use(func(c *gin.Context) {
		c.Set(constants.UserIdKey, 1)
		c.Set(constants.RolesKey, []interface{}{"user", "admin"})
		c.Next()
	})
	r.Use(Authorization([]string{"admin"}))
	r.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "welcome admin")
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "welcome admin", w.Body.String())
}

func TestAuthorization_UserHasOneOfMultipleValidRoles(t *testing.T) {
	r, w := setupGin()

	r.Use(func(c *gin.Context) {
		c.Set(constants.UserIdKey, 42)
		c.Set(constants.RolesKey, []interface{}{"manager"})
		c.Next()
	})
	r.Use(Authorization([]string{"admin", "manager", "supervisor"}))
	r.GET("/dashboard", func(c *gin.Context) {
		c.String(http.StatusOK, "welcome manager")
	})

	req, _ := http.NewRequest(http.MethodGet, "/dashboard", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "welcome manager", w.Body.String())
}
