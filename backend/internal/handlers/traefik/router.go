package traefik

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/models"
	"gorm.io/gorm"
)

type RouterHandler struct {
	db *gorm.DB
}

func NewRouterHandler(db *gorm.DB) *RouterHandler {
	return &RouterHandler{db: db}
}

// ListRouters returns list of all routers
func (h *RouterHandler) ListRouters(c *gin.Context) {
	var routers []models.Router

	if err := h.db.Preload("Hostnames").
		Preload("Service").
		Preload("Middlewares.Middleware").
		Find(&routers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch routers"})
		return
	}

	// Convert to response format
	responses := make([]models.RouterResponse, len(routers))
	for i, router := range routers {
		responses[i] = router.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{"routers": responses})
}

// GetRouter returns a specific router
func (h *RouterHandler) GetRouter(c *gin.Context) {
	id := c.Param("id")

	routerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid router ID"})
		return
	}

	var router models.Router
	if err := h.db.Preload("Hostnames").
		Preload("Service").
		Preload("Middlewares.Middleware").
		First(&router, routerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}

	c.JSON(http.StatusOK, router.ToResponse())
}

// CreateRouter creates a new router
func (h *RouterHandler) CreateRouter(c *gin.Context) {
	var req models.CreateRouterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if name already exists
	var existingRouter models.Router
	if err := h.db.Where("name = ?", req.Name).First(&existingRouter).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Router with this name already exists"})
		return
	}

	// Validate service exists
	var service models.Service
	if err := h.db.First(&service, req.ServiceID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service not found"})
		return
	}

	// Validate hostnames
	for _, hostname := range req.Hostnames {
		if !isValidHostname(hostname) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hostname: " + hostname})
			return
		}
	}

	// Validate middlewares if provided
	if len(req.MiddlewareIDs) > 0 {
		var middlewareCount int64
		h.db.Model(&models.Middleware{}).Where("id IN ?", req.MiddlewareIDs).Count(&middlewareCount)
		if int(middlewareCount) != len(req.MiddlewareIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more middlewares not found"})
			return
		}
	}

	// Set default entry points
	entryPoints := "web,websecure"
	if len(req.EntryPoints) > 0 {
		entryPoints = strings.Join(req.EntryPoints, ",")
	}

	// Set default cert resolver
	certResolver := "letsencrypt"
	if req.TLSCertResolver != "" {
		certResolver = req.TLSCertResolver
	}

	// Create router
	router := models.Router{
		Name:            req.Name,
		ServiceID:       req.ServiceID,
		TLSEnabled:      req.TLSEnabled,
		TLSCertResolver: certResolver,
		RedirectHTTPS:   req.RedirectHTTPS,
		EntryPoints:     entryPoints,
		IsActive:        true,
	}

	// Create hostnames
	hostnames := make([]models.RouterHostname, len(req.Hostnames))
	for i, hostname := range req.Hostnames {
		hostnames[i] = models.RouterHostname{
			Hostname: hostname,
		}
	}
	router.Hostnames = hostnames

	// Create router first to get ID
	if err := h.db.Create(&router).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create router"})
		return
	}

	// Associate middlewares if provided
	if len(req.MiddlewareIDs) > 0 {
		routerMiddlewares := make([]models.RouterMiddleware, len(req.MiddlewareIDs))
		for i, midID := range req.MiddlewareIDs {
			routerMiddlewares[i] = models.RouterMiddleware{
				RouterID:     router.ID,
				MiddlewareID: midID,
				Priority:     i,
			}
		}
		h.db.Create(&routerMiddlewares)
	}

	// Reload with associations
	h.db.Preload("Hostnames").
		Preload("Service").
		Preload("Middlewares.Middleware").
		First(&router, router.ID)

	c.JSON(http.StatusCreated, router.ToResponse())
}

// UpdateRouter updates a router
func (h *RouterHandler) UpdateRouter(c *gin.Context) {
	id := c.Param("id")

	routerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid router ID"})
		return
	}

	var router models.Router
	if err := h.db.Preload("Hostnames").
		Preload("Middlewares").
		First(&router, routerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}

	var req models.UpdateRouterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update name if provided
	if req.Name != nil && *req.Name != "" {
		// Check if name is taken by another router
		var existingRouter models.Router
		if err := h.db.Where("name = ? AND id != ?", *req.Name, routerID).First(&existingRouter).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Router name already in use"})
			return
		}
		router.Name = *req.Name
	}

	// Update service if provided
	if req.ServiceID != nil {
		var service models.Service
		if err := h.db.First(&service, *req.ServiceID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Service not found"})
			return
		}
		router.ServiceID = *req.ServiceID
	}

	// Update hostnames if provided
	if req.Hostnames != nil && len(req.Hostnames) > 0 {
		// Validate hostnames
		for _, hostname := range req.Hostnames {
			if !isValidHostname(hostname) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hostname: " + hostname})
				return
			}
		}

		// Delete old hostnames
		h.db.Where("router_id = ?", router.ID).Delete(&models.RouterHostname{})

		// Create new hostnames
		hostnames := make([]models.RouterHostname, len(req.Hostnames))
		for i, hostname := range req.Hostnames {
			hostnames[i] = models.RouterHostname{
				RouterID: router.ID,
				Hostname: hostname,
			}
		}
		router.Hostnames = hostnames
	}

	// Update TLS settings
	if req.TLSEnabled != nil {
		router.TLSEnabled = *req.TLSEnabled
	}
	if req.TLSCertResolver != nil {
		router.TLSCertResolver = *req.TLSCertResolver
	}
	if req.RedirectHTTPS != nil {
		router.RedirectHTTPS = *req.RedirectHTTPS
	}
	if req.IsActive != nil {
		router.IsActive = *req.IsActive
	}

	// Update entry points if provided
	if req.EntryPoints != nil && len(req.EntryPoints) > 0 {
		router.EntryPoints = strings.Join(req.EntryPoints, ",")
	}

	// Update middlewares if provided
	if req.MiddlewareIDs != nil {
		// Validate middlewares
		if len(req.MiddlewareIDs) > 0 {
			var middlewareCount int64
			h.db.Model(&models.Middleware{}).Where("id IN ?", req.MiddlewareIDs).Count(&middlewareCount)
			if int(middlewareCount) != len(req.MiddlewareIDs) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "One or more middlewares not found"})
				return
			}
		}

		// Delete old associations
		h.db.Where("router_id = ?", router.ID).Delete(&models.RouterMiddleware{})

		// Create new associations
		routerMiddlewares := make([]models.RouterMiddleware, len(req.MiddlewareIDs))
		for i, midID := range req.MiddlewareIDs {
			routerMiddlewares[i] = models.RouterMiddleware{
				RouterID:     router.ID,
				MiddlewareID: midID,
				Priority:     i,
			}
		}
		h.db.Create(&routerMiddlewares)
	}

	// Save router
	if err := h.db.Save(&router).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update router"})
		return
	}

	// Reload with associations
	h.db.Preload("Hostnames").
		Preload("Service").
		Preload("Middlewares.Middleware").
		First(&router, router.ID)

	c.JSON(http.StatusOK, router.ToResponse())
}

// DeleteRouter deletes a router
func (h *RouterHandler) DeleteRouter(c *gin.Context) {
	id := c.Param("id")

	routerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid router ID"})
		return
	}

	var router models.Router
	if err := h.db.First(&router, routerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}

	// Delete hostnames
	h.db.Where("router_id = ?", router.ID).Delete(&models.RouterHostname{})

	// Delete middleware associations
	h.db.Where("router_id = ?", router.ID).Delete(&models.RouterMiddleware{})

	// Delete router
	if err := h.db.Delete(&router).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete router"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Router deleted successfully"})
}

// Helper function to validate hostnames
func isValidHostname(hostname string) bool {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return false
	}
	// Basic hostname validation
	if strings.Contains(hostname, " ") {
		return false
	}
	// Must contain at least one dot (domain.tld)
	if !strings.Contains(hostname, ".") {
		return false
	}
	return true
}
