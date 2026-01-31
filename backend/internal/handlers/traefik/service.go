package traefik

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/models"
	"gorm.io/gorm"
)

type ServiceHandler struct {
	db *gorm.DB
}

func NewServiceHandler(db *gorm.DB) *ServiceHandler {
	return &ServiceHandler{db: db}
}

// ListServices returns list of all services
func (h *ServiceHandler) ListServices(c *gin.Context) {
	var services []models.Service

	if err := h.db.Preload("Servers").Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services"})
		return
	}

	// Convert to response format
	responses := make([]models.ServiceResponse, len(services))
	for i, service := range services {
		responses[i] = service.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{"services": responses})
}

// GetService returns a specific service
func (h *ServiceHandler) GetService(c *gin.Context) {
	id := c.Param("id")

	serviceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.Preload("Servers").First(&service, serviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, service.ToResponse())
}

// CreateService creates a new service
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req models.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if name already exists
	var existingService models.Service
	if err := h.db.Where("name = ?", req.Name).First(&existingService).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Service with this name already exists"})
		return
	}

	// Validate server URLs
	for _, url := range req.Servers {
		if !isValidServerURL(url) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server URL: " + url})
			return
		}
	}

	// Set default load balancer type
	lbType := req.LoadBalancerType
	if lbType == "" {
		lbType = "wrr"
	}

	// Create service
	service := models.Service{
		Name:               req.Name,
		Type:               "http",
		LoadBalancerType:   lbType,
		PassHostHeader:     req.PassHostHeader,
		HealthCheckEnabled: req.HealthCheckEnabled,
		HealthCheckPath:    req.HealthCheckPath,
		IsActive:           true,
	}

	if req.HealthCheckInterval > 0 {
		service.HealthCheckInterval = req.HealthCheckInterval
	} else {
		service.HealthCheckInterval = 10
	}

	// Create servers
	servers := make([]models.ServiceServer, len(req.Servers))
	for i, url := range req.Servers {
		servers[i] = models.ServiceServer{
			URL:    url,
			Weight: 1,
		}
	}
	service.Servers = servers

	if err := h.db.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	// Reload with servers
	h.db.Preload("Servers").First(&service, service.ID)

	c.JSON(http.StatusCreated, service.ToResponse())
}

// UpdateService updates a service
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")

	serviceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.Preload("Servers").First(&service, serviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	var req models.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update name if provided
	if req.Name != nil && *req.Name != "" {
		// Check if name is taken by another service
		var existingService models.Service
		if err := h.db.Where("name = ? AND id != ?", *req.Name, serviceID).First(&existingService).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Service name already in use"})
			return
		}
		service.Name = *req.Name
	}

	// Update other fields
	if req.LoadBalancerType != nil {
		service.LoadBalancerType = *req.LoadBalancerType
	}
	if req.PassHostHeader != nil {
		service.PassHostHeader = *req.PassHostHeader
	}
	if req.HealthCheckEnabled != nil {
		service.HealthCheckEnabled = *req.HealthCheckEnabled
	}
	if req.HealthCheckPath != nil {
		service.HealthCheckPath = *req.HealthCheckPath
	}
	if req.HealthCheckInterval != nil {
		service.HealthCheckInterval = *req.HealthCheckInterval
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	// Update servers if provided
	if req.Servers != nil && len(req.Servers) > 0 {
		// Validate server URLs
		for _, url := range req.Servers {
			if !isValidServerURL(url) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server URL: " + url})
				return
			}
		}

		// Delete old servers
		h.db.Where("service_id = ?", service.ID).Delete(&models.ServiceServer{})

		// Create new servers
		servers := make([]models.ServiceServer, len(req.Servers))
		for i, url := range req.Servers {
			servers[i] = models.ServiceServer{
				ServiceID: service.ID,
				URL:       url,
				Weight:    1,
			}
		}
		service.Servers = servers
	}

	if err := h.db.Save(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	// Reload with servers
	h.db.Preload("Servers").First(&service, service.ID)

	c.JSON(http.StatusOK, service.ToResponse())
}

// DeleteService deletes a service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")

	serviceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.First(&service, serviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Check if service is in use by any routers
	var routerCount int64
	h.db.Model(&models.Router{}).Where("service_id = ?", serviceID).Count(&routerCount)
	if routerCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete service: it is used by one or more routers"})
		return
	}

	// Delete servers first
	h.db.Where("service_id = ?", service.ID).Delete(&models.ServiceServer{})

	// Delete service
	if err := h.db.Delete(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}

// Helper function to validate server URLs
func isValidServerURL(url string) bool {
	url = strings.TrimSpace(url)
	if url == "" {
		return false
	}
	// Must start with http:// or https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}
	return true
}

type MiddlewareHandler struct {
	db *gorm.DB
}

func NewMiddlewareHandler(db *gorm.DB) *MiddlewareHandler {
	return &MiddlewareHandler{db: db}
}

// ListMiddlewares returns list of all middlewares
func (h *MiddlewareHandler) ListMiddlewares(c *gin.Context) {
	var middlewares []models.Middleware

	if err := h.db.Find(&middlewares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch middlewares"})
		return
	}

	// Convert to response format
	responses := make([]models.MiddlewareResponse, len(middlewares))
	for i, m := range middlewares {
		responses[i] = m.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{"middlewares": responses})
}

// GetMiddleware returns a specific middleware
func (h *MiddlewareHandler) GetMiddleware(c *gin.Context) {
	id := c.Param("id")

	middlewareID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid middleware ID"})
		return
	}

	var middleware models.Middleware
	if err := h.db.First(&middleware, middlewareID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Middleware not found"})
		return
	}

	c.JSON(http.StatusOK, middleware.ToResponse())
}

// CreateMiddleware creates a new middleware
func (h *MiddlewareHandler) CreateMiddleware(c *gin.Context) {
	var req models.CreateMiddlewareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if name already exists
	var existingMiddleware models.Middleware
	if err := h.db.Where("name = ?", req.Name).First(&existingMiddleware).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Middleware with this name already exists"})
		return
	}

	// Validate and serialize config
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid middleware configuration"})
		return
	}

	// Create middleware
	middleware := models.Middleware{
		Name:     req.Name,
		Type:     req.Type,
		Config:   string(configJSON),
		IsActive: true,
	}

	if err := h.db.Create(&middleware).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create middleware"})
		return
	}

	c.JSON(http.StatusCreated, middleware.ToResponse())
}

// UpdateMiddleware updates a middleware
func (h *MiddlewareHandler) UpdateMiddleware(c *gin.Context) {
	id := c.Param("id")

	middlewareID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid middleware ID"})
		return
	}

	var middleware models.Middleware
	if err := h.db.First(&middleware, middlewareID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Middleware not found"})
		return
	}

	var req models.UpdateMiddlewareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update name if provided
	if req.Name != nil && *req.Name != "" {
		// Check if name is taken by another middleware
		var existingMiddleware models.Middleware
		if err := h.db.Where("name = ? AND id != ?", *req.Name, middlewareID).First(&existingMiddleware).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Middleware name already in use"})
			return
		}
		middleware.Name = *req.Name
	}

	// Update type if provided
	if req.Type != nil && *req.Type != "" {
		middleware.Type = *req.Type
	}

	// Update config if provided
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid middleware configuration"})
			return
		}
		middleware.Config = string(configJSON)
	}

	// Update is_active if provided
	if req.IsActive != nil {
		middleware.IsActive = *req.IsActive
	}

	if err := h.db.Save(&middleware).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update middleware"})
		return
	}

	c.JSON(http.StatusOK, middleware.ToResponse())
}

// DeleteMiddleware deletes a middleware
func (h *MiddlewareHandler) DeleteMiddleware(c *gin.Context) {
	id := c.Param("id")

	middlewareID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid middleware ID"})
		return
	}

	var middleware models.Middleware
	if err := h.db.First(&middleware, middlewareID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Middleware not found"})
		return
	}

	// Check if middleware is in use by any routers
	var routerMiddlewareCount int64
	h.db.Model(&models.RouterMiddleware{}).Where("middleware_id = ?", middlewareID).Count(&routerMiddlewareCount)
	if routerMiddlewareCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete middleware: it is used by one or more routers"})
		return
	}

	// Delete middleware
	if err := h.db.Delete(&middleware).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete middleware"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Middleware deleted successfully"})
}
