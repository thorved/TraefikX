package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/config"
	"github.com/traefikx/backend/internal/handlers"
	"github.com/traefikx/backend/internal/middleware"
	"github.com/traefikx/backend/internal/routes/auth"
	"github.com/traefikx/backend/internal/routes/static"
	traefikRoutes "github.com/traefikx/backend/internal/routes/traefik"
	"github.com/traefikx/backend/internal/routes/user"
	"github.com/traefikx/backend/internal/services"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, aggregator *services.AggregatorService) *gin.Engine {
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
		auth.RegisterRoutes(api, authHandler)
		user.RegisterRoutes(api, userHandler)

		// Traefik routes
		traefikRoutes.RegisterRoutes(api, cfg, db, aggregator)
	}

	// Static routes
	static.RegisterRoutes(r)

	return r
}
