package traefik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/models"
	"github.com/traefikx/backend/internal/services"
	"gorm.io/gorm"
)

type HTTPProviderHandler struct {
	db         *gorm.DB
	aggregator *services.AggregatorService
}

func NewHTTPProviderHandler(db *gorm.DB, aggregator *services.AggregatorService) *HTTPProviderHandler {
	return &HTTPProviderHandler{
		db:         db,
		aggregator: aggregator,
	}
}

// ListHTTPProviders returns all HTTP providers
func (h *HTTPProviderHandler) ListHTTPProviders(c *gin.Context) {
	var providers []models.HTTPProvider

	if err := h.db.Order("priority DESC, created_at ASC").Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}

	// Convert to response format
	responses := make([]map[string]interface{}, len(providers))
	for i, provider := range providers {
		responses[i] = provider.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{"providers": responses})
}

// GetHTTPProvider returns a specific provider
func (h *HTTPProviderHandler) GetHTTPProvider(c *gin.Context) {
	id := c.Param("id")

	providerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider models.HTTPProvider
	if err := h.db.First(&provider, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	c.JSON(http.StatusOK, provider.ToResponse())
}

// CreateHTTPProvider creates a new HTTP provider
func (h *HTTPProviderHandler) CreateHTTPProvider(c *gin.Context) {
	var req models.CreateHTTPProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if name already exists
	var existingProvider models.HTTPProvider
	if err := h.db.Where("name = ?", req.Name).First(&existingProvider).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Provider with this name already exists"})
		return
	}

	// Set default refresh interval
	refreshInterval := req.RefreshInterval
	if refreshInterval < 5 {
		refreshInterval = 30 // Default 30 seconds
	}

	provider := models.HTTPProvider{
		Name:            req.Name,
		URL:             req.URL,
		Priority:        req.Priority,
		IsActive:        req.IsActive,
		RefreshInterval: refreshInterval,
	}

	if err := h.db.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}

	// Start polling if active
	if provider.IsActive && h.aggregator != nil {
		h.aggregator.AddProvider(&provider)
	}

	c.JSON(http.StatusCreated, provider.ToResponse())
}

// UpdateHTTPProvider updates a provider
func (h *HTTPProviderHandler) UpdateHTTPProvider(c *gin.Context) {
	id := c.Param("id")

	providerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider models.HTTPProvider
	if err := h.db.First(&provider, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	var req models.UpdateHTTPProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update name if provided
	if req.Name != nil && *req.Name != "" {
		// Check if name is taken by another provider
		var existingProvider models.HTTPProvider
		if err := h.db.Where("name = ? AND id != ?", *req.Name, providerID).First(&existingProvider).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Provider name already in use"})
			return
		}
		provider.Name = *req.Name
	}

	// Update URL if provided
	if req.URL != nil && *req.URL != "" {
		provider.URL = *req.URL
	}

	// Update priority if provided
	if req.Priority != nil {
		provider.Priority = *req.Priority
	}

	// Update refresh interval if provided
	if req.RefreshInterval != nil && *req.RefreshInterval >= 5 {
		provider.RefreshInterval = *req.RefreshInterval
	}

	// Update active status if provided
	if req.IsActive != nil {
		wasActive := provider.IsActive
		provider.IsActive = *req.IsActive

		// Handle activation/deactivation
		if h.aggregator != nil {
			if *req.IsActive && !wasActive {
				// Activating
				h.aggregator.AddProvider(&provider)
			} else if !*req.IsActive && wasActive {
				// Deactivating
				h.aggregator.DeleteProvider(provider.ID)
			}
		}
	}

	if err := h.db.Save(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}

	// Restart polling if config changed and still active
	if h.aggregator != nil && provider.IsActive {
		h.aggregator.UpdateProvider(&provider)
	}

	c.JSON(http.StatusOK, provider.ToResponse())
}

// DeleteHTTPProvider deletes a provider
func (h *HTTPProviderHandler) DeleteHTTPProvider(c *gin.Context) {
	id := c.Param("id")

	providerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider models.HTTPProvider
	if err := h.db.First(&provider, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Stop polling
	if h.aggregator != nil {
		h.aggregator.DeleteProvider(provider.ID)
	}

	// Delete from database
	if err := h.db.Delete(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}

// RefreshHTTPProvider manually triggers a refresh for a provider
func (h *HTTPProviderHandler) RefreshHTTPProvider(c *gin.Context) {
	id := c.Param("id")

	providerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	if h.aggregator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Aggregator service not available"})
		return
	}

	if err := h.aggregator.RefreshProvider(uint(providerID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Refresh triggered"})
}

// TestHTTPProvider tests the connection to a provider
func (h *HTTPProviderHandler) TestHTTPProvider(c *gin.Context) {
	id := c.Param("id")

	providerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider models.HTTPProvider
	if err := h.db.First(&provider, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Trigger a refresh and wait a moment
	if h.aggregator != nil {
		h.aggregator.RefreshProvider(uint(providerID))
	}

	// Reload to get latest status
	h.db.First(&provider, providerID)

	c.JSON(http.StatusOK, provider.ToResponse())
}

// GetMergedConfig returns the merged configuration from all providers
func (h *HTTPProviderHandler) GetMergedConfig(c *gin.Context) {
	if h.aggregator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Aggregator service not available"})
		return
	}

	// Get local config from database
	var routers []models.Router
	h.db.Preload("Hostnames").Preload("Service").Preload("Middlewares.Middleware").Find(&routers)

	var services []models.Service
	h.db.Preload("Servers").Find(&services)

	var middlewares []models.Middleware
	h.db.Find(&middlewares)

	// Convert to Traefik format
	localRouters := make(map[string]interface{})
	for _, router := range routers {
		if !router.IsActive {
			continue
		}
		localRouters[router.Name] = convertRouterToTraefikFormat(&router)
	}

	localServices := make(map[string]interface{})
	for _, service := range services {
		if !service.IsActive {
			continue
		}
		localServices[service.Name] = convertServiceToTraefikFormat(&service)
	}

	localMiddlewares := make(map[string]interface{})
	for _, middleware := range middlewares {
		if !middleware.IsActive {
			continue
		}
		localMiddlewares[middleware.Name] = convertMiddlewareToTraefikFormat(&middleware)
	}

	// Get merged config
	mergedConfig, conflicts := h.aggregator.GetMergedConfig(localRouters, localServices, localMiddlewares)

	c.JSON(http.StatusOK, gin.H{
		"config":    mergedConfig,
		"conflicts": conflicts,
		"sources":   h.getSourcesInfo(),
	})
}

// getSourcesInfo returns information about all provider sources
func (h *HTTPProviderHandler) getSourcesInfo() []gin.H {
	var providers []models.HTTPProvider
	h.db.Order("priority DESC").Find(&providers)

	// Count local items
	var localRouterCount, localServiceCount, localMiddlewareCount int64
	h.db.Model(&models.Router{}).Where("is_active = ?", true).Count(&localRouterCount)
	h.db.Model(&models.Service{}).Where("is_active = ?", true).Count(&localServiceCount)
	h.db.Model(&models.Middleware{}).Where("is_active = ?", true).Count(&localMiddlewareCount)

	sources := []gin.H{
		{
			"name":             "local",
			"priority":         9999, // Local always has highest priority
			"status":           "healthy",
			"router_count":     localRouterCount,
			"service_count":    localServiceCount,
			"middleware_count": localMiddlewareCount,
		},
	}

	for _, provider := range providers {
		status := "healthy"
		if provider.LastError != "" {
			if provider.LastFetched == nil {
				status = "unhealthy"
			} else {
				status = "degraded"
			}
		} else if !provider.IsActive {
			status = "inactive"
		}

		sources = append(sources, gin.H{
			"name":             provider.Name,
			"priority":         provider.Priority,
			"status":           status,
			"last_fetched":     provider.LastFetched,
			"last_error":       provider.LastError,
			"router_count":     provider.RouterCount,
			"service_count":    provider.ServiceCount,
			"middleware_count": provider.MiddlewareCount,
		})
	}

	return sources
}

// Helper functions to convert models to Traefik format
func convertRouterToTraefikFormat(router *models.Router) map[string]interface{} {
	if len(router.Hostnames) == 0 {
		return nil
	}

	// Build rule from hostnames
	rule := buildRuleFromHostnames(router.Hostnames)
	entryPoints := splitEntryPoints(router.EntryPoints)

	middlewareNames := []string{}
	if router.RedirectHTTPS {
		middlewareNames = append(middlewareNames, fmt.Sprintf("%s-redirect-https", router.Name))
	}
	for _, rm := range router.Middlewares {
		if rm.Middleware.IsActive {
			middlewareNames = append(middlewareNames, rm.Middleware.Name)
		}
	}

	result := map[string]interface{}{
		"entryPoints": entryPoints,
		"rule":        rule,
		"service":     router.Service.Name,
	}

	if len(middlewareNames) > 0 {
		result["middlewares"] = middlewareNames
	}

	if router.TLSEnabled {
		result["tls"] = map[string]interface{}{
			"certResolver": router.TLSCertResolver,
		}
	}

	return result
}

func buildRuleFromHostnames(hostnames []models.RouterHostname) string {
	if len(hostnames) == 0 {
		return ""
	}
	if len(hostnames) == 1 {
		return fmt.Sprintf("Host(`%s`)", hostnames[0].Hostname)
	}
	hosts := make([]string, len(hostnames))
	for i, h := range hostnames {
		hosts[i] = fmt.Sprintf("`%s`", h.Hostname)
	}
	return fmt.Sprintf("Host(%s)", strings.Join(hosts, ", "))
}

func convertServiceToTraefikFormat(service *models.Service) map[string]interface{} {
	servers := make([]map[string]string, len(service.Servers))
	for i, s := range service.Servers {
		servers[i] = map[string]string{"url": s.URL}
	}

	lbConfig := map[string]interface{}{
		"servers":        servers,
		"passHostHeader": service.PassHostHeader,
	}

	if service.HealthCheckEnabled && service.HealthCheckPath != "" {
		lbConfig["healthCheck"] = map[string]interface{}{
			"path":     service.HealthCheckPath,
			"interval": fmt.Sprintf("%ds", service.HealthCheckInterval),
		}
	}

	return map[string]interface{}{
		"loadBalancer": lbConfig,
	}
}

func convertMiddlewareToTraefikFormat(middleware *models.Middleware) map[string]interface{} {
	var config models.MiddlewareConfig
	if err := json.Unmarshal([]byte(middleware.Config), &config); err != nil {
		return nil
	}

	switch middleware.Type {
	case "redirectScheme":
		return map[string]interface{}{
			"redirectScheme": map[string]interface{}{
				"scheme":    config.Scheme,
				"port":      config.Port,
				"permanent": config.Permanent,
			},
		}
	case "headers":
		result := map[string]interface{}{}
		if len(config.CustomRequestHeaders) > 0 {
			result["customRequestHeaders"] = config.CustomRequestHeaders
		}
		if len(config.CustomResponseHeaders) > 0 {
			result["customResponseHeaders"] = config.CustomResponseHeaders
		}
		if config.SSLRedirect {
			result["sslRedirect"] = true
		}
		return map[string]interface{}{"headers": result}
	case "stripPrefix":
		return map[string]interface{}{
			"stripPrefix": map[string]interface{}{
				"prefixes":   config.Prefixes,
				"forceSlash": config.ForceSlash,
			},
		}
	case "addPrefix":
		return map[string]interface{}{
			"addPrefix": map[string]interface{}{
				"prefix": config.Prefix,
			},
		}
	default:
		return nil
	}
}
