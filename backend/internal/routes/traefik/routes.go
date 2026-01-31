package traefik

import (
	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/config"
	"github.com/traefikx/backend/internal/handlers/traefik"
	"github.com/traefikx/backend/internal/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(api *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {
	// Initialize handlers
	serviceHandler := traefik.NewServiceHandler(db)
	routerHandler := traefik.NewRouterHandler(db)
	middlewareHandler := traefik.NewMiddlewareHandler(db)
	providerHandler := traefik.NewTraefikProviderHandler(db)
	proxyHandler := traefik.NewProxyHandler(db)

	// Traefik management routes (protected)
	traefikGroup := api.Group("/traefik")
	traefikGroup.Use(middleware.AuthMiddleware())
	{
		// Proxy Hosts - Available to all authenticated users (with ownership checks)
		traefikGroup.GET("/proxies", proxyHandler.ListProxyHosts)
		traefikGroup.POST("/proxies", proxyHandler.CreateProxyHost)
		traefikGroup.GET("/proxies/:id", proxyHandler.GetProxyHost)
		traefikGroup.PUT("/proxies/:id", proxyHandler.UpdateProxyHost)
		traefikGroup.DELETE("/proxies/:id", proxyHandler.DeleteProxyHost)

		// Service management (admin only)
		traefikGroup.GET("/services", middleware.AdminMiddleware(), serviceHandler.ListServices)
		traefikGroup.POST("/services", middleware.AdminMiddleware(), serviceHandler.CreateService)
		traefikGroup.GET("/services/:id", middleware.AdminMiddleware(), serviceHandler.GetService)
		traefikGroup.PUT("/services/:id", middleware.AdminMiddleware(), serviceHandler.UpdateService)
		traefikGroup.DELETE("/services/:id", middleware.AdminMiddleware(), serviceHandler.DeleteService)

		// Middleware management (admin only)
		traefikGroup.GET("/middlewares", middleware.AdminMiddleware(), middlewareHandler.ListMiddlewares)
		traefikGroup.POST("/middlewares", middleware.AdminMiddleware(), middlewareHandler.CreateMiddleware)
		traefikGroup.GET("/middlewares/:id", middleware.AdminMiddleware(), middlewareHandler.GetMiddleware)
		traefikGroup.PUT("/middlewares/:id", middleware.AdminMiddleware(), middlewareHandler.UpdateMiddleware)
		traefikGroup.DELETE("/middlewares/:id", middleware.AdminMiddleware(), middlewareHandler.DeleteMiddleware)

		// Router management (admin only)
		traefikGroup.GET("/routers", middleware.AdminMiddleware(), routerHandler.ListRouters)
		traefikGroup.POST("/routers", middleware.AdminMiddleware(), routerHandler.CreateRouter)
		traefikGroup.GET("/routers/:id", middleware.AdminMiddleware(), routerHandler.GetRouter)
		traefikGroup.PUT("/routers/:id", middleware.AdminMiddleware(), routerHandler.UpdateRouter)
		traefikGroup.DELETE("/routers/:id", middleware.AdminMiddleware(), routerHandler.DeleteRouter)
	}

	// Traefik provider endpoint (public but token-protected)
	api.GET("/traefik/provider/config", middleware.TraefikProviderTokenMiddleware(cfg.TraefikProviderToken), providerHandler.GenerateConfig)
}
