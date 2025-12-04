// internal/config/cors.go
package config

import (
	"os"
	"strings"
)

// GetCORSOrigins returns allowed origins from environment variable
func GetCORSOrigins() []string {
	originsStr := os.Getenv("CORS_ALLOWED_ORIGINS")

	// Default origins if not set
	if originsStr == "" {
		return []string{
			"http://localhost:3000",
			"http://localhost:8080",
		}
	}

	// Split comma-separated origins
	origins := strings.Split(originsStr, ",")

	// Trim whitespace from each origin
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}
