package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	// Set Gin to Test Mode to disable debug logs
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		errPayload     any
		expectedStatus int
	}{
		{
			name:           "Scenario 1: Standard error instance",
			errPayload:     errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Scenario 2: String error payload (non-error type)",
			errPayload:     "something unexpected happened",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Scenario 3: Custom struct payload (non-error type)",
			errPayload:     struct{ Code int }{Code: 500},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Scenario 4: Primitive integer payload (non-error type)",
			errPayload:     404,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Scenario 5: Nil payload",
			errPayload:     nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin context with response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Execute ErrorHandler
			ErrorHandler(c, tt.errPayload)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.True(t, c.IsAborted(), "Context should be aborted")
			assert.NotEmpty(t, w.Body.String(), "Response body should not be empty")
		})
	}
}
