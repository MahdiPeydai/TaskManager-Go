package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLimitByRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		rateLimit      float64
		requests       int
		expectedStatus int
	}{
		{
			name:           "request within rate limit",
			rateLimit:      1,
			requests:       1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "request exceeds rate limit",
			rateLimit:      1,
			requests:       2,
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			nextCalled := 0

			router.Use(LimitByRequest(tt.rateLimit))

			router.GET("/", func(c *gin.Context) {
				nextCalled++
				c.Status(http.StatusOK)
			})

			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("X-Real-IP", "192.168.1.10")

				rec := httptest.NewRecorder()

				router.ServeHTTP(rec, req)

				if i == tt.requests-1 {
					assert.Equal(t, tt.expectedStatus, rec.Code)
				}
			}

			if tt.expectedStatus == http.StatusTooManyRequests {
				assert.Equal(t, 1, nextCalled)
			}
		})
	}
}

func TestLimitByRequest_UsesRealIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.Use(LimitByRequest(1))

	var receivedIP string

	router.GET("/", func(c *gin.Context) {
		receivedIP = c.GetHeader("X-Real-IP")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, receivedIP)
}

func TestLimitByRequest_TooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.Use(LimitByRequest(1))

	nextCalled := 0

	router.GET("/", func(c *gin.Context) {
		nextCalled++
		c.Status(http.StatusOK)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "192.168.1.10")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if i == 1 {
			assert.Equal(t, http.StatusTooManyRequests, rec.Code)
			assert.Equal(t, 1, nextCalled)
		}
	}
}
