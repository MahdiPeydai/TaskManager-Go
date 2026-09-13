package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHealthEndpoint_Get tests the health check endpoint
func TestHealthEndpoint_Get(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	tests := []struct {
		name           string
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "successful health check",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.NotNil(t, body)
				assert.True(t, body["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("GET", "/api/v1/health", nil)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}
