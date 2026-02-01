package traefik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefikx/backend/internal/models"
	"github.com/traefikx/backend/internal/services"
	"gorm.io/gorm"
)

// TraefikProviderHandler generates dynamic configuration for Traefik
type TraefikProviderHandler struct {
	db         *gorm.DB
	aggregator *services.AggregatorService
}

func NewTraefikProviderHandler(db *gorm.DB, aggregator *services.AggregatorService) *TraefikProviderHandler {
	return &TraefikProviderHandler{db: db, aggregator: aggregator}
}

// GenerateConfig generates the dynamic configuration for Traefik
// This merges local configuration with external endpoint configurations
// Priority: Local (highest) > External endpoints (by priority)
func (h *TraefikProviderHandler) GenerateConfig(c *gin.Context) {
	// Initialize config with official Traefik types - wrap HTTP config properly
	config := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Services:          make(map[string]*dynamic.Service),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Models:            make(map[string]*dynamic.Model),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
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

	// Fetch all active servers transports
	// Note: You may need to add a ServersTransport model if you want to store these in DB
	// For now, we'll assume they come from HTTP providers only

	// Track local items for conflict detection
	localRouters := make(map[string]*dynamic.Router)
	localServices := make(map[string]*dynamic.Service)
	localMiddlewares := make(map[string]*dynamic.Middleware)
	localServersTransports := make(map[string]*dynamic.ServersTransport)

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

		// TLS config removed for testing
		// Build TLS config with domains
		var tlsConfig *dynamic.RouterTLSConfig
		/* if router.TLSEnabled {
			tlsConfig = &dynamic.RouterTLSConfig{
				CertResolver: router.TLSCertResolver,
			}

			// Add domains from hostnames if TLS is enabled
			if len(router.Hostnames) > 0 {
				hostnames := make([]string, len(router.Hostnames))
				for i, h := range router.Hostnames {
					hostnames[i] = h.Hostname
				}

				if len(hostnames) > 0 {
					mainDomain := hostnames[0]
					sans := []string{}

					// Add wildcard for the main domain
					parts := strings.Split(mainDomain, ".")
					if len(parts) >= 2 {
						wildcard := "*." + strings.Join(parts[1:], ".")
						sans = append(sans, wildcard)
					}

					// Add any additional hostnames as SANs
					for i := 1; i < len(hostnames); i++ {
						sans = append(sans, hostnames[i])
						parts := strings.Split(hostnames[i], ".")
						if len(parts) >= 2 {
							wildcard := "*." + strings.Join(parts[1:], ".")
							sans = append(sans, wildcard)
						}
					}

					tlsConfig.Domains = []types.Domain{
						{
							Main: mainDomain,
							SANs: sans,
						},
					}
				}
			}
		} */

		// Add router to config
		dynRouter := &dynamic.Router{
			EntryPoints: entryPoints,
			Rule:        rule,
			Service:     router.Service.Name,
			Middlewares: middlewareNames,
			TLS:         tlsConfig,
		}
		config.HTTP.Routers[router.Name] = dynRouter
		localRouters[router.Name] = dynRouter

		// Add service to config
		if _, exists := config.HTTP.Services[router.Service.Name]; !exists {
			serviceConfig := buildServiceConfig(&router.Service)
			config.HTTP.Services[router.Service.Name] = serviceConfig
			localServices[router.Service.Name] = serviceConfig
		}
	}

	// Generate middleware configs
	for _, middleware := range middlewares {
		middlewareConfig := buildMiddlewareConfig(&middleware)
		if middlewareConfig != nil {
			config.HTTP.Middlewares[middleware.Name] = middlewareConfig
			localMiddlewares[middleware.Name] = middlewareConfig
		}
	}

	// Generate redirect-to-https middlewares for routers that need it
	for _, router := range routers {
		if router.RedirectHTTPS {
			middlewareName := fmt.Sprintf("%s-redirect-https", router.Name)
			mw := &dynamic.Middleware{
				RedirectScheme: &dynamic.RedirectScheme{
					Scheme:    "https",
					Port:      "443",
					Permanent: true,
				},
			}
			config.HTTP.Middlewares[middlewareName] = mw
			localMiddlewares[middlewareName] = mw
		}
	}

	// Merge external endpoint configurations (if aggregator is available)
	if h.aggregator != nil {
		mergedHTTP := h.mergeExternalConfigs(config.HTTP, localRouters, localServices, localMiddlewares, localServersTransports)
		config.HTTP = mergedHTTP
	}

	// Return full configuration wrapped with "http" key
	c.JSON(http.StatusOK, config)
}

// mergeExternalConfigs merges configurations from external endpoints
func (h *TraefikProviderHandler) mergeExternalConfigs(
	config *dynamic.HTTPConfiguration,
	localRouters map[string]*dynamic.Router,
	localServices map[string]*dynamic.Service,
	localMiddlewares map[string]*dynamic.Middleware,
	localServersTransports map[string]*dynamic.ServersTransport,
) *dynamic.HTTPConfiguration {
	// Get merged config from aggregator using official types
	mergedConfig, conflicts := h.aggregator.GetMergedConfig(
		localRouters,
		localServices,
		localMiddlewares,
		localServersTransports,
	)

	// Log conflicts for debugging
	for _, conflict := range conflicts {
		fmt.Printf("Config conflict: %s '%s' from '%s' overridden by '%s' (priority: %d)\n",
			conflict.Type, conflict.Name, conflict.Source, conflict.OverriddenBy, conflict.SourcePriority)
	}

	// Return the merged HTTP configuration
	if mergedConfig != nil && mergedConfig.HTTP != nil {
		return mergedConfig.HTTP
	}

	return config
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

func buildServiceConfig(service *models.Service) *dynamic.Service {
	servers := make([]dynamic.Server, 0, len(service.Servers))
	for _, s := range service.Servers {
		servers = append(servers, dynamic.Server{
			URL: s.URL,
		})
	}

	passHostHeader := service.PassHostHeader
	config := &dynamic.Service{
		LoadBalancer: &dynamic.ServersLoadBalancer{
			Servers:        servers,
			PassHostHeader: &passHostHeader,
		},
	}

	if service.HealthCheckEnabled && service.HealthCheckPath != "" {
		config.LoadBalancer.HealthCheck = &dynamic.ServerHealthCheck{
			Path: service.HealthCheckPath,
		}
	}

	return config
}

func buildMiddlewareConfig(middleware *models.Middleware) *dynamic.Middleware {
	var config models.MiddlewareConfig
	if err := json.Unmarshal([]byte(middleware.Config), &config); err != nil {
		// Invalid config - skip this middleware
		return nil
	}

	switch middleware.Type {
	case "redirectScheme":
		return &dynamic.Middleware{
			RedirectScheme: &dynamic.RedirectScheme{
				Scheme:    config.Scheme,
				Port:      config.Port,
				Permanent: config.Permanent,
			},
		}
	case "headers":
		return &dynamic.Middleware{
			Headers: &dynamic.Headers{
				CustomRequestHeaders:  config.CustomRequestHeaders,
				CustomResponseHeaders: config.CustomResponseHeaders,
			},
		}
	case "stripPrefix":
		return &dynamic.Middleware{
			StripPrefix: &dynamic.StripPrefix{
				Prefixes: config.Prefixes,
			},
		}
	case "addPrefix":
		return &dynamic.Middleware{
			AddPrefix: &dynamic.AddPrefix{
				Prefix: config.Prefix,
			},
		}
	default:
		return nil
	}
}
