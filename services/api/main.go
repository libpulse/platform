package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
	"github.com/libpulse/platform/services/api/internal/auth"
	"github.com/libpulse/platform/services/api/internal/config"
	"github.com/libpulse/platform/services/api/internal/handlers"
	"github.com/libpulse/platform/services/api/internal/supabase"
)

type Config struct {
	JWTSecret      []byte
	ServiceRoleKey string
	AuthBaseURL    string
}

func loadConfigFromEnv() (*Config, error) {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	serviceRole := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	authURL := os.Getenv("SUPABASE_AUTH_URL")

	if jwtSecret == "" || serviceRole == "" || authURL == "" {
		return nil, ErrMissingEnv
	}

	return &Config{
		JWTSecret:      []byte(jwtSecret),
		ServiceRoleKey: serviceRole,
		AuthBaseURL:    authURL,
	}, nil
}

var ErrMissingEnv = &configError{"SUPABASE_JWT_SECRET, SUPABASE_SERVICE_ROLE_KEY, SUPABASE_AUTH_URL must be set"}

type configError struct{ msg string }

func (e *configError) Error() string { return e.msg }

func main() {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Supabase Admin API client
	sbClient := supabase.NewClient(cfg.AuthBaseURL, cfg.ServiceRoleKey)

	// Gin router
	r := gin.Default()

	// Get CORS origins from environment
	corsOrigins := config.GetCORSOrigins()
	log.Printf("CORS allowed origins: %v", corsOrigins)

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Protected API routes
	api := r.Group("/api/v1")
	// Create Adapter Store
	userStore := &supabase.UserStore{Client: sbClient}

	api.Use(auth.NewMiddleware(cfg.JWTSecret))
	{
		api.GET("/me", handlers.GetCurrentUserHandler(userStore))
	}

	addr := ":8080"
	log.Printf("LibPulse API listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
