package env

import "testing"

func TestIsSecret(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"DATABASE_PASSWORD", true},
		{"API_KEY", true},
		{"STRIPE_SECRET_KEY", true},
		{"JWT_TOKEN", true},
		{"PRIVATE_KEY", true},
		{"AWS_ACCESS_KEY", true},
		{"AUTH_TOKEN", true},
		{"DB_CREDENTIAL", true},
		{"DATABASE_DSN", true},
		{"DATABASE_URL", false},
		{"DEBUG", false},
		{"NODE_ENV", false},
		{"PORT", false},
		{"HOST", false},
		{"app_secret", true}, // case insensitive
		{"ApiKey", true},     // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := IsSecret(tt.key)
			if got != tt.expected {
				t.Errorf("IsSecret(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}
