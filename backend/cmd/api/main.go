package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/auth"
	"github.com/traefikx/backend/internal/config"
	"github.com/traefikx/backend/internal/database"
	"github.com/traefikx/backend/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create default admin user if no users exist
	if err := database.CreateDefaultAdmin(cfg); err != nil {
		log.Fatalf("Failed to create default admin: %v", err)
	}

	// Initialize OIDC if enabled
	if cfg.OIDCEnabled {
		if err := auth.InitOIDC(cfg); err != nil {
			log.Printf("Warning: Failed to initialize OIDC: %v", err)
		} else {
			log.Println("OIDC initialized successfully")
		}
	}

	// Setup router
	r := routes.SetupRouter(cfg, db)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
