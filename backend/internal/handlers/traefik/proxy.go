package traefik

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/models"
	"gorm.io/gorm"
)

// ProxyHost represents a combined router + service for the UI
type ProxyHost struct {
	ID            uint     `json:"id"`
	DomainNames   []string `json:"domain_names"`
	ForwardScheme string   `json:"forward_scheme"` // http or https
	ForwardHost   string   `json:"forward_host"`
	ForwardPort   int      `json:"forward_port"`
	SSL           bool     `json:"ssl"`
	SSLProvider   string   `json:"ssl_provider,omitempty"`
	Access        string   `json:"access"` // public, private
	Status        string   `json:"status"` // online, offline
	CreatedAt     string   `json:"created_at"`
}

// CreateProxyHostRequest for creating a new proxy
type CreateProxyHostRequest struct {
	DomainNames   []string `json:"domain_names" binding:"required,min=1"`
	ForwardScheme string   `json:"forward_scheme" binding:"required,oneof=http https"`
	ForwardHost   string   `json:"forward_host" binding:"required"`
	ForwardPort   int      `json:"forward_port" binding:"required,min=1,max=65535"`
	SSL           bool     `json:"ssl"`
	SSLProvider   string   `json:"ssl_provider,omitempty"`
	Access        string   `json:"access" binding:"required,oneof=public private"`
}

// UpdateProxyHostRequest for updating a proxy
type UpdateProxyHostRequest struct {
	DomainNames   []string `json:"domain_names,omitempty"`
	ForwardScheme string   `json:"forward_scheme,omitempty" binding:"omitempty,oneof=http https"`
	ForwardHost   string   `json:"forward_host,omitempty"`
	ForwardPort   int      `json:"forward_port,omitempty" binding:"omitempty,min=1,max=65535"`
	SSL           *bool    `json:"ssl,omitempty"`
	SSLProvider   *string  `json:"ssl_provider,omitempty"`
	Access        string   `json:"access,omitempty" binding:"omitempty,oneof=public private"`
}

type ProxyHandler struct {
	db *gorm.DB
}

func NewProxyHandler(db *gorm.DB) *ProxyHandler {
	return &ProxyHandler{db: db}
}

// ListProxyHosts returns all proxy hosts (combined router + service view)
// Admin sees all, regular users see only their own
func (h *ProxyHandler) ListProxyHosts(c *gin.Context) {
	// Get user info from context
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	var routers []models.Router
	query := h.db.Where("is_active = ?", true).
		Preload("Hostnames").
		Preload("Service.Servers")

	// If not admin, filter by user_id
	if role != string(models.RoleAdmin) {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&routers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch proxy hosts"})
		return
	}

	proxies := make([]ProxyHost, 0, len(routers))
	for _, router := range routers {
		proxy := h.routerToProxyHost(&router)
		proxies = append(proxies, proxy)
	}

	c.JSON(http.StatusOK, gin.H{"proxies": proxies})
}

// GetProxyHost returns a single proxy host
// Users can only access their own, admins can access all
func (h *ProxyHandler) GetProxyHost(c *gin.Context) {
	id := c.Param("id")
	routerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	// Get user info from context
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	query := h.db.Where("id = ? AND is_active = ?", routerID, true).
		Preload("Hostnames").
		Preload("Service.Servers")

	// If not admin, filter by user_id
	if role != string(models.RoleAdmin) {
		query = query.Where("user_id = ?", userID)
	}

	var router models.Router
	if err := query.First(&router).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy host not found"})
		return
	}

	proxy := h.routerToProxyHost(&router)
	c.JSON(http.StatusOK, proxy)
}

// CreateProxyHost creates a new proxy (router + service together)
func (h *ProxyHandler) CreateProxyHost(c *gin.Context) {
	var req CreateProxyHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate names from first domain
	if len(req.DomainNames) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one domain name is required"})
		return
	}

	sanitizedDomain := sanitizeName(req.DomainNames[0])
	serviceName := fmt.Sprintf("%s-service", sanitizedDomain)
	routerName := fmt.Sprintf("%s-router", sanitizedDomain)

	// Check if router name already exists
	var existingRouter models.Router
	if err := h.db.Where("name = ?", routerName).First(&existingRouter).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "A proxy for this domain already exists"})
		return
	}

	// Build forward URL
	forwardURL := fmt.Sprintf("%s://%s:%d", req.ForwardScheme, req.ForwardHost, req.ForwardPort)

	// Create service
	service := models.Service{
		Name:             serviceName,
		Type:             "http",
		LoadBalancerType: "wrr",
		PassHostHeader:   true,
		IsActive:         true,
		Servers: []models.ServiceServer{
			{URL: forwardURL, Weight: 1},
		},
	}

	if err := h.db.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	// Create router
	sslProvider := ""
	if req.SSL {
		sslProvider = "letsencrypt"
		if req.SSLProvider != "" {
			sslProvider = req.SSLProvider
		}
	}

	// Get user ID from context
	userID, _ := c.Get("userID")

	router := models.Router{
		Name:            routerName,
		ServiceID:       service.ID,
		UserID:          userID.(uint),
		TLSEnabled:      req.SSL,
		TLSCertResolver: sslProvider,
		RedirectHTTPS:   req.SSL, // Auto-redirect to HTTPS when SSL is enabled
		EntryPoints:     "web,websecure",
		IsActive:        true,
	}

	// Create hostnames
	hostnames := make([]models.RouterHostname, len(req.DomainNames))
	for i, domain := range req.DomainNames {
		hostnames[i] = models.RouterHostname{Hostname: domain}
	}
	router.Hostnames = hostnames

	if err := h.db.Create(&router).Error; err != nil {
		// Cleanup service on router failure
		h.db.Delete(&service)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create router"})
		return
	}

	// Reload with associations
	h.db.Preload("Hostnames").Preload("Service.Servers").First(&router, router.ID)

	proxy := h.routerToProxyHost(&router)
	c.JSON(http.StatusCreated, proxy)
}

// UpdateProxyHost updates a proxy
// Users can only update their own, admins can update all
func (h *ProxyHandler) UpdateProxyHost(c *gin.Context) {
	id := c.Param("id")
	routerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	// Get user info from context
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	query := h.db.Where("id = ? AND is_active = ?", routerID, true).
		Preload("Hostnames").
		Preload("Service.Servers")

	// If not admin, filter by user_id
	if role != string(models.RoleAdmin) {
		query = query.Where("user_id = ?", userID)
	}

	var router models.Router
	if err := query.First(&router).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy host not found or access denied"})
		return
	}

	var req UpdateProxyHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update hostnames
	if req.DomainNames != nil && len(req.DomainNames) > 0 {
		// Delete old hostnames
		if err := h.db.Where("router_id = ?", router.ID).Delete(&models.RouterHostname{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove old domains"})
			return
		}

		// Clear old hostnames from memory
		router.Hostnames = nil

		// Create new hostnames one by one to ensure proper creation
		for _, domain := range req.DomainNames {
			hostname := models.RouterHostname{
				RouterID: router.ID,
				Hostname: domain,
			}
			if err := h.db.Create(&hostname).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add domain: " + domain})
				return
			}
		}
	}

	// Update service
	if req.ForwardScheme != "" || req.ForwardHost != "" || req.ForwardPort > 0 {
		scheme := req.ForwardScheme
		host := req.ForwardHost
		port := req.ForwardPort

		// Use existing values if not provided
		if scheme == "" {
			scheme = h.getSchemeFromURL(router.Service.Servers[0].URL)
		}
		if host == "" {
			host = h.getHostFromURL(router.Service.Servers[0].URL)
		}
		if port == 0 {
			port = h.getPortFromURL(router.Service.Servers[0].URL)
		}

		forwardURL := fmt.Sprintf("%s://%s:%d", scheme, host, port)

		// Delete old servers
		if err := h.db.Where("service_id = ?", router.ServiceID).Delete(&models.ServiceServer{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove old server"})
			return
		}

		// Clear old servers from memory
		router.Service.Servers = nil

		// Create new server
		newServer := models.ServiceServer{
			ServiceID: router.ServiceID,
			URL:       forwardURL,
			Weight:    1,
		}
		if err := h.db.Create(&newServer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add server"})
			return
		}
	}

	// Update SSL
	if req.SSL != nil {
		router.TLSEnabled = *req.SSL
		router.RedirectHTTPS = *req.SSL
		if *req.SSL {
			router.TLSCertResolver = "letsencrypt"
			if req.SSLProvider != nil && *req.SSLProvider != "" {
				router.TLSCertResolver = *req.SSLProvider
			}
		} else {
			router.TLSCertResolver = ""
		}
	}

	if err := h.db.Save(&router).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proxy"})
		return
	}

	// Reload
	h.db.Preload("Hostnames").Preload("Service.Servers").First(&router, router.ID)
	proxy := h.routerToProxyHost(&router)
	c.JSON(http.StatusOK, proxy)
}

// DeleteProxyHost deletes a proxy (router + service)
// Users can only delete their own, admins can delete all
func (h *ProxyHandler) DeleteProxyHost(c *gin.Context) {
	id := c.Param("id")
	routerID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	// Get user info from context
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	query := h.db.Where("id = ?", routerID)

	// If not admin, filter by user_id
	if role != string(models.RoleAdmin) {
		query = query.Where("user_id = ?", userID)
	}

	var router models.Router
	if err := query.First(&router).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy host not found or access denied"})
		return
	}

	serviceID := router.ServiceID

	// Delete router and related data
	h.db.Where("router_id = ?", routerID).Delete(&models.RouterHostname{})
	h.db.Where("router_id = ?", routerID).Delete(&models.RouterMiddleware{})
	h.db.Delete(&router)

	// Delete service and servers
	h.db.Where("service_id = ?", serviceID).Delete(&models.ServiceServer{})
	h.db.Delete(&models.Service{}, serviceID)

	c.JSON(http.StatusOK, gin.H{"message": "Proxy host deleted successfully"})
}

// Helper functions
func (h *ProxyHandler) routerToProxyHost(router *models.Router) ProxyHost {
	domains := make([]string, len(router.Hostnames))
	for i, h := range router.Hostnames {
		domains[i] = h.Hostname
	}

	// Parse forward URL
	var forwardScheme, forwardHost string
	var forwardPort int
	if len(router.Service.Servers) > 0 {
		url := router.Service.Servers[0].URL
		forwardScheme = h.getSchemeFromURL(url)
		forwardHost = h.getHostFromURL(url)
		forwardPort = h.getPortFromURL(url)
	}

	sslProvider := ""
	if router.TLSEnabled {
		sslProvider = router.TLSCertResolver
		if sslProvider == "" {
			sslProvider = "Let's Encrypt"
		}
	}

	access := "public"
	// TODO: Check if any access middleware is applied

	status := "online"
	if !router.IsActive {
		status = "offline"
	}

	return ProxyHost{
		ID:            router.ID,
		DomainNames:   domains,
		ForwardScheme: forwardScheme,
		ForwardHost:   forwardHost,
		ForwardPort:   forwardPort,
		SSL:           router.TLSEnabled,
		SSLProvider:   sslProvider,
		Access:        access,
		Status:        status,
		CreatedAt:     router.CreatedAt.Format("Jan 2, 2006, 3:04 PM"),
	}
}

func (h *ProxyHandler) getSchemeFromURL(url string) string {
	if strings.HasPrefix(url, "https://") {
		return "https"
	}
	return "http"
}

func (h *ProxyHandler) getHostFromURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func (h *ProxyHandler) getPortFromURL(url string) int {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, ":")
	if len(parts) > 1 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}
	// Default ports
	if strings.HasPrefix(url, "https") {
		return 443
	}
	return 80
}

func sanitizeName(name string) string {
	// Replace dots and special chars with dashes
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)
	return name
}
