package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/handlers"
	"github.com/traefikx/backend/internal/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, handler *handlers.AuthHandler) {
	// Public routes
	api.POST("/auth/login", handler.Login)
	api.POST("/auth/refresh", handler.Refresh)
	api.GET("/auth/oidc", handler.OIDCLogin)
	api.GET("/auth/oidc/callback", handler.OIDCCallback)
	api.GET("/auth/oidc/status", handler.GetOIDCStatus)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/auth/logout", handler.Logout)
		protected.GET("/auth/me", handler.GetMe)
		protected.PUT("/auth/password", handler.ChangePassword)
		protected.POST("/auth/password/toggle", handler.TogglePasswordLogin)
		protected.DELETE("/auth/password", handler.RemovePassword)
		protected.POST("/auth/oidc/link", handler.OIDCLinkInit)
		protected.DELETE("/auth/oidc/link", handler.OIDCUnlink)
	}
}
