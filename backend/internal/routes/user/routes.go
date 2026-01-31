package user

import (
	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/handlers"
	"github.com/traefikx/backend/internal/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, handler *handlers.UserHandler) {
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// These routes require admin privileges
		protected.GET("/users", middleware.AdminMiddleware(), handler.ListUsers)
		protected.POST("/users", middleware.AdminMiddleware(), handler.CreateUser)
		protected.PUT("/users/:id", middleware.AdminMiddleware(), handler.UpdateUser)
		protected.DELETE("/users/:id", middleware.AdminMiddleware(), handler.DeleteUser)
		protected.POST("/users/:id/reset-password", middleware.AdminMiddleware(), handler.ResetPassword)
		protected.POST("/users/:id/password/toggle", middleware.AdminMiddleware(), handler.ToggleUserPasswordLogin)
		protected.POST("/users/:id/oidc/toggle", middleware.AdminMiddleware(), handler.ToggleUserOIDC)

		// This route is accessible by authenticated users (handler might have its own checks)
		protected.GET("/users/:id", handler.GetUser)
	}
}
