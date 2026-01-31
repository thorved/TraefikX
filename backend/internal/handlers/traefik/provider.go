package traefik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/models"
	"gorm.io/gorm"
)

// TraefikProviderHandler generates dynamic configuration for Traefik
type TraefikProviderHandler struct {
	db *gorm.DB
}

func NewTraefikProviderHandler(db *gorm.DB) *TraefikProviderHandler {
	return &TraefikProviderHandler{db: db}
}

// TraefikDynamicConfig represents the full dynamic configuration
type TraefikDynamicConfig struct {
	HTTP *HTTPConfig `json:"http,omitempty"`
}

type HTTPConfig struct {
	Routers     map[string]*RouterConfig     `json:"routers,omitempty"`
	Services    map[string]*ServiceConfig    `json:"services,omitempty"`
	Middlewares map[string]*MiddlewareConfig `json:"middlewares,omitempty"`
}

type RouterConfig struct {
	EntryPoints []string   `json:"entryPoints,omitempty"`
	Rule        string     `json:"rule,omitempty"`
	Service     string     `json:"service,omitempty"`
	Middlewares []string   `json:"middlewares,omitempty"`
	TLS         *TLSConfig `json:"tls,omitempty"`
}

type TLSConfig struct {
	CertResolver string `json:"certResolver,omitempty"`
}

type ServiceConfig struct {
	LoadBalancer *LoadBalancerConfig `json:"loadBalancer,omitempty"`
}

type LoadBalancerConfig struct {
	Servers        []ServerConfig `json:"servers,omitempty"`
	PassHostHeader bool           `json:"passHostHeader,omitempty"`
	HealthCheck    *HealthCheck   `json:"healthCheck,omitempty"`
}

type ServerConfig struct {
	URL string `json:"url,omitempty"`
}

type HealthCheck struct {
	Path     string `json:"path,omitempty"`
	Interval string `json:"interval,omitempty"`
}

// GenerateConfig generates the dynamic configuration for Traefik
func (h *TraefikProviderHandler) GenerateConfig(c *gin.Context) {
	config := &TraefikDynamicConfig{
		HTTP: &HTTPConfig{
			Routers:     make(map[string]*RouterConfig),
			Services:    make(map[string]*ServiceConfig),
			Middlewares: make(map[string]*MiddlewareConfig),
		},
	}

	// Fetch all active routers with their associations
	var routers []models.Router
	if err := h.db.Where("is_active = ?", true).
		Preload("Hostnames").
		Preload("Service.Servers").
		Preload("Middlewares.Middleware").
		Find(&routers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch routers"})
		return
	}

	// Fetch all active middlewares
	var middlewares []models.Middleware
	if err := h.db.Where("is_active = ?", true).Find(&middlewares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch middlewares"})
		return
	}

	// Generate router and service configs
	for _, router := range routers {
		// Skip if no hostnames
		if len(router.Hostnames) == 0 {
			continue
		}

		// Build rule from hostnames
		rule := buildRule(router.Hostnames)

		// Build entry points
		entryPoints := splitEntryPoints(router.EntryPoints)

		// Build middleware list
		middlewareNames := []string{}
		if router.RedirectHTTPS {
			// Add redirect-to-https middleware automatically
			middlewareNames = append(middlewareNames, fmt.Sprintf("%s-redirect-https", router.Name))
		}
		for _, rm := range router.Middlewares {
			if rm.Middleware.IsActive {
				middlewareNames = append(middlewareNames, rm.Middleware.Name)
			}
		}

		// Build TLS config
		var tlsConfig *TLSConfig
		if router.TLSEnabled {
			tlsConfig = &TLSConfig{
				CertResolver: router.TLSCertResolver,
			}
		}

		// Add router to config
		config.HTTP.Routers[router.Name] = &RouterConfig{
			EntryPoints: entryPoints,
			Rule:        rule,
			Service:     router.Service.Name,
			Middlewares: middlewareNames,
			TLS:         tlsConfig,
		}

		// Add service to config
		if _, exists := config.HTTP.Services[router.Service.Name]; !exists {
			serviceConfig := buildServiceConfig(&router.Service)
			config.HTTP.Services[router.Service.Name] = serviceConfig
		}
	}

	// Generate middleware configs
	for _, middleware := range middlewares {
		middlewareConfig := buildMiddlewareConfig(&middleware)
		if middlewareConfig != nil {
			config.HTTP.Middlewares[middleware.Name] = middlewareConfig
		}
	}

	// Generate redirect-to-https middlewares for routers that need it
	for _, router := range routers {
		if router.RedirectHTTPS {
			middlewareName := fmt.Sprintf("%s-redirect-https", router.Name)
			config.HTTP.Middlewares[middlewareName] = &MiddlewareConfig{
				RedirectScheme: &RedirectSchemeConfig{
					Scheme:    "https",
					Port:      "443",
					Permanent: true,
				},
			}
		}
	}

	c.JSON(http.StatusOK, config)
}

// Helper functions for building Traefik config

func splitEntryPoints(ep string) []string {
	if ep == "" {
		return []string{"web", "websecure"}
	}
	return splitAndTrim(ep)
}

func splitAndTrim(s string) []string {
	parts := []string{}
	for _, p := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func buildRule(hostnames []models.RouterHostname) string {
	if len(hostnames) == 0 {
		return ""
	}
	if len(hostnames) == 1 {
		return fmt.Sprintf("Host(`%s`)", hostnames[0].Hostname)
	}
	// Multiple hostnames - use Host()
	hosts := make([]string, len(hostnames))
	for i, h := range hostnames {
		hosts[i] = fmt.Sprintf("`%s`", h.Hostname)
	}
	return fmt.Sprintf("Host(%s)", strings.Join(hosts, ", "))
}

func buildServiceConfig(service *models.Service) *ServiceConfig {
	servers := make([]ServerConfig, 0, len(service.Servers))
	for _, s := range service.Servers {
		servers = append(servers, ServerConfig{
			URL: s.URL,
		})
	}

	config := &ServiceConfig{
		LoadBalancer: &LoadBalancerConfig{
			Servers:        servers,
			PassHostHeader: service.PassHostHeader,
		},
	}

	if service.HealthCheckEnabled && service.HealthCheckPath != "" {
		config.LoadBalancer.HealthCheck = &HealthCheck{
			Path:     service.HealthCheckPath,
			Interval: fmt.Sprintf("%ds", service.HealthCheckInterval),
		}
	}

	return config
}

// MiddlewareConfig for Traefik - supports different middleware types
type MiddlewareConfigOutput struct {
	RedirectScheme *RedirectSchemeConfig `json:"redirectScheme,omitempty"`
	Headers        *HeadersConfig        `json:"headers,omitempty"`
	StripPrefix    *StripPrefixConfig    `json:"stripPrefix,omitempty"`
	AddPrefix      *AddPrefixConfig      `json:"addPrefix,omitempty"`
}

// Alias for the output struct (same fields)
type MiddlewareConfig = MiddlewareConfigOutput

type RedirectSchemeConfig struct {
	Scheme    string `json:"scheme,omitempty"`
	Port      string `json:"port,omitempty"`
	Permanent bool   `json:"permanent,omitempty"`
}

type HeadersConfig struct {
	CustomRequestHeaders  map[string]string `json:"customRequestHeaders,omitempty"`
	CustomResponseHeaders map[string]string `json:"customResponseHeaders,omitempty"`
	SSLRedirect           bool              `json:"sslRedirect,omitempty"`
}

type StripPrefixConfig struct {
	Prefixes   []string `json:"prefixes,omitempty"`
	ForceSlash bool     `json:"forceSlash,omitempty"`
}

type AddPrefixConfig struct {
	Prefix string `json:"prefix,omitempty"`
}

func buildMiddlewareConfig(middleware *models.Middleware) *MiddlewareConfigOutput {
	var config models.MiddlewareConfig
	if err := json.Unmarshal([]byte(middleware.Config), &config); err != nil {
		// Invalid config - skip this middleware
		return nil
	}

	switch middleware.Type {
	case "redirectScheme":
		return &MiddlewareConfigOutput{
			RedirectScheme: &RedirectSchemeConfig{
				Scheme:    config.Scheme,
				Port:      config.Port,
				Permanent: config.Permanent,
			},
		}
	case "headers":
		return &MiddlewareConfigOutput{
			Headers: &HeadersConfig{
				CustomRequestHeaders:  config.CustomRequestHeaders,
				CustomResponseHeaders: config.CustomResponseHeaders,
				SSLRedirect:           config.SSLRedirect,
			},
		}
	case "stripPrefix":
		return &MiddlewareConfigOutput{
			StripPrefix: &StripPrefixConfig{
				Prefixes:   config.Prefixes,
				ForceSlash: config.ForceSlash,
			},
		}
	case "addPrefix":
		return &MiddlewareConfigOutput{
			AddPrefix: &AddPrefixConfig{
				Prefix: config.Prefix,
			},
		}
	default:
		return nil
	}
}
