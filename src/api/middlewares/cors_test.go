package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const allowOrigins = "http://localhost:3000"

	tests := []struct {
		name               string
		method             string
		expectedStatus     int
		expectedNextCalled bool
	}{
		{
			name:               "options request",
			method:             http.MethodOptions,
			expectedStatus:     http.StatusNoContent,
			expectedNextCalled: false,
		},
		{
			name:               "get request",
			method:             http.MethodGet,
			expectedStatus:     http.StatusOK,
			expectedNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			nextCalled := false

			router.Use(Cors(allowOrigins))

			router.GET("/", func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedNextCalled, nextCalled)

			assert.Equal(
				t,
				allowOrigins,
				rec.Header().Get("Access-Control-Allow-Origin"),
			)

			assert.Equal(
				t,
				"true",
				rec.Header().Get("Access-Control-Allow-Credentials"),
			)

			assert.Equal(
				t,
				"21600",
				rec.Header().Get("Access-Control-Max-Age"),
			)

			assert.Equal(
				t,
				"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
				rec.Header().Get("Access-Control-Allow-Headers"),
			)

			assert.Equal(
				t,
				"GET, POST, PUT, DELETE, OPTIONS, PATCH",
				rec.Header().Get("Access-Control-Allow-Methods"),
			)
		})
	}
}
