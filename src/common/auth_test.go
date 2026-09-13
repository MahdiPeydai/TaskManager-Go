package common

import (
	"testing"

	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/stretchr/testify/assert"
)

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{
			name:     "admin role",
			roles:    []string{constants.AdminRoleName},
			expected: true,
		},
		{
			name:     "default role",
			roles:    []string{"default"},
			expected: false,
		},
		{
			name: "admin among multiple roles",
			roles: []string{
				"default",
				constants.AdminRoleName,
			},
			expected: true,
		},
		{
			name:     "empty roles",
			roles:    []string{},
			expected: false,
		},
		{
			name:     "unrelated roles",
			roles:    []string{"manager", "developer"},
			expected: false,
		},
		{
			name:     "nil roles",
			roles:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAdmin(tt.roles))
		})
	}
}
