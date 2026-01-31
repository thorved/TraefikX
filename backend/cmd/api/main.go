package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/auth"
	"github.com/traefikx/backend/internal/config"
	"github.com/traefikx/backend/internal/database"
	"github.com/traefikx/backend/internal/handlers"
	"github.com/traefikx/backend/internal/middleware"
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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	userHandler := handlers.NewUserHandler(db)

	// Setup router
	r := gin.Default()

	// CORS middleware
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)
		api.GET("/auth/oidc", authHandler.OIDCLogin)
		api.GET("/auth/oidc/callback", authHandler.OIDCCallback)
		api.GET("/auth/oidc/status", authHandler.GetOIDCStatus)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Auth routes
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.GetMe)
			protected.PUT("/auth/password", authHandler.ChangePassword)
			protected.POST("/auth/password/toggle", authHandler.TogglePasswordLogin)
			protected.DELETE("/auth/password", authHandler.RemovePassword)
			protected.POST("/auth/oidc/link", authHandler.OIDCLinkInit)
			protected.DELETE("/auth/oidc/link", authHandler.OIDCUnlink)

			// User routes
			protected.GET("/users", middleware.AdminMiddleware(), userHandler.ListUsers)
			protected.POST("/users", middleware.AdminMiddleware(), userHandler.CreateUser)
			protected.GET("/users/:id", userHandler.GetUser)
			protected.PUT("/users/:id", middleware.AdminMiddleware(), userHandler.UpdateUser)
			protected.DELETE("/users/:id", middleware.AdminMiddleware(), userHandler.DeleteUser)
			protected.POST("/users/:id/reset-password", middleware.AdminMiddleware(), userHandler.ResetPassword)
			protected.POST("/users/:id/password/toggle", middleware.AdminMiddleware(), userHandler.ToggleUserPasswordLogin)
			protected.POST("/users/:id/oidc/toggle", middleware.AdminMiddleware(), userHandler.ToggleUserOIDC)
		}
	}

	// Serve frontend static files
	staticPath := getStaticPath()
	if staticPath != "" {
		log.Printf("Serving static files from: %s", staticPath)
		// Check if static directory exists
		if _, err := os.Stat(staticPath); !os.IsNotExist(err) {
			// Serve Next.js static files
			r.Static("/_next", filepath.Join(staticPath, "_next"))
			r.StaticFile("/favicon.ico", filepath.Join(staticPath, "favicon.ico"))

			// Serve page directories (login, dashboard, etc.)
			r.Static("/login", filepath.Join(staticPath, "login"))
			r.Static("/dashboard", filepath.Join(staticPath, "dashboard"))
			r.Static("/users", filepath.Join(staticPath, "users"))
			r.Static("/profile", filepath.Join(staticPath, "profile"))
			r.Static("/auth", filepath.Join(staticPath, "auth"))

			// Serve index.html for root path
			r.GET("/", func(c *gin.Context) {
				c.File(filepath.Join(staticPath, "index.html"))
			})

			// Serve index.html for all non-API, non-static routes (SPA fallback)
			r.NoRoute(func(c *gin.Context) {
				path := c.Request.URL.Path
				if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/_next/") {
					c.File(filepath.Join(staticPath, "index.html"))
				} else {
					c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				}
			})
		} else {
			log.Printf("Warning: Static path not found: %s", staticPath)
		}
	} else {
		log.Println("Warning: No static files path configured")
	}

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

func getStaticPath() string {
	// Try multiple possible locations for the Next.js frontend build
	// Next.js with output: 'export' creates an 'out' directory
	possiblePaths := []string{
		// Same directory as binary (production)
		func() string {
			execPath, err := os.Executable()
			if err != nil {
				return ""
			}
			return filepath.Join(filepath.Dir(execPath), "frontend", "out")
		}(),
		// Parent directory of backend (development from backend/)
		"../frontend/out",
		// Current working directory
		"./frontend/out",
		// Absolute path from project root
		"./../frontend/out",
	}

	for _, path := range possiblePaths {
		if path == "" {
			continue
		}
		// Check if index.html exists in this path
		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return path
		}
	}

	return ""
}
