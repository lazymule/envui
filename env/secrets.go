package env

import "strings"

var secretPatterns = []string{
	"PASSWORD",
	"SECRET",
	"TOKEN",
	"API_KEY",
	"APIKEY",
	"API_SECRET",
	"PRIVATE_KEY",
	"ACCESS_KEY",
	"AUTH",
	"CREDENTIAL",
	"DSN", // Database connection strings often have passwords
}

// IsSecret checks if a key name looks like it might contain a secret
func IsSecret(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range secretPatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}
