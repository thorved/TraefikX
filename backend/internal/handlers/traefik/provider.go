package traefik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	Priority    int        `json:"priority,omitempty"`
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
// This merges local configuration with external endpoint configurations
// Priority: Local (highest) > External endpoints (by priority)
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

	// Track local items for conflict detection
	localRouters := make(map[string]bool)
	localServices := make(map[string]bool)
	localMiddlewares := make(map[string]bool)

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
		localRouters[router.Name] = true

		// Add service to config
		if _, exists := config.HTTP.Services[router.Service.Name]; !exists {
			serviceConfig := buildServiceConfig(&router.Service)
			config.HTTP.Services[router.Service.Name] = serviceConfig
			localServices[router.Service.Name] = true
		}
	}

	// Generate middleware configs
	for _, middleware := range middlewares {
		middlewareConfig := buildMiddlewareConfig(&middleware)
		if middlewareConfig != nil {
			config.HTTP.Middlewares[middleware.Name] = middlewareConfig
			localMiddlewares[middleware.Name] = true
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
			localMiddlewares[middlewareName] = true
		}
	}

	// Merge external endpoint configurations (if aggregator is available)
	if h.aggregator != nil {
		h.mergeExternalConfigs(config, localRouters, localServices, localMiddlewares)
	}

	c.JSON(http.StatusOK, config)
}

// mergeExternalConfigs merges configurations from external endpoints
func (h *TraefikProviderHandler) mergeExternalConfigs(config *TraefikDynamicConfig, localRouters, localServices, localMiddlewares map[string]bool) {
	// Convert local configs to interface{} maps for aggregator
	localRoutersMap := make(map[string]interface{})
	localServicesMap := make(map[string]interface{})
	localMiddlewaresMap := make(map[string]interface{})

	for name := range localRouters {
		if router, exists := config.HTTP.Routers[name]; exists {
			localRoutersMap[name] = convertRouterConfigToMap(router)
		}
	}
	for name := range localServices {
		if service, exists := config.HTTP.Services[name]; exists {
			localServicesMap[name] = convertServiceConfigToMap(service)
		}
	}
	for name := range localMiddlewares {
		if middleware, exists := config.HTTP.Middlewares[name]; exists {
			localMiddlewaresMap[name] = convertMiddlewareConfigToMap(middleware)
		}
	}

	// Get merged config from aggregator
	mergedConfig, conflicts := h.aggregator.GetMergedConfig(localRoutersMap, localServicesMap, localMiddlewaresMap)

	// Log conflicts for debugging
	for _, conflict := range conflicts {
		fmt.Printf("Config conflict: %s '%s' from '%s' overridden by '%s' (priority: %d)\n",
			conflict.Type, conflict.Name, conflict.Source, conflict.OverriddenBy, conflict.SourcePriority)
	}

	// Merge external routers (skip if local already has it)
	for name, router := range mergedConfig.HTTP.Routers {
		if _, exists := localRouters[name]; !exists {
			if routerMap, ok := router.(map[string]interface{}); ok {
				config.HTTP.Routers[name] = convertMapToRouterConfig(routerMap)
			}
		}
	}

	// Merge external services
	for name, service := range mergedConfig.HTTP.Services {
		if _, exists := localServices[name]; !exists {
			if serviceMap, ok := service.(map[string]interface{}); ok {
				config.HTTP.Services[name] = convertMapToServiceConfig(serviceMap)
			}
		}
	}

	// Merge external middlewares
	for name, middleware := range mergedConfig.HTTP.Middlewares {
		if _, exists := localMiddlewares[name]; !exists {
			if middlewareMap, ok := middleware.(map[string]interface{}); ok {
				config.HTTP.Middlewares[name] = convertMapToMiddlewareConfig(middlewareMap)
			}
		}
	}
}

// Helper functions for converting between config types and maps
func convertRouterConfigToMap(router *RouterConfig) map[string]interface{} {
	result := map[string]interface{}{
		"entryPoints": router.EntryPoints,
		"rule":        router.Rule,
		"service":     router.Service,
	}
	if len(router.Middlewares) > 0 {
		result["middlewares"] = router.Middlewares
	}
	if router.TLS != nil {
		result["tls"] = map[string]interface{}{
			"certResolver": router.TLS.CertResolver,
		}
	}
	if router.Priority > 0 {
		result["priority"] = router.Priority
	}
	return result
}

func convertServiceConfigToMap(service *ServiceConfig) map[string]interface{} {
	if service.LoadBalancer == nil {
		return nil
	}
	servers := make([]map[string]string, len(service.LoadBalancer.Servers))
	for i, s := range service.LoadBalancer.Servers {
		servers[i] = map[string]string{"url": s.URL}
	}
	lbConfig := map[string]interface{}{
		"servers":        servers,
		"passHostHeader": service.LoadBalancer.PassHostHeader,
	}
	if service.LoadBalancer.HealthCheck != nil {
		lbConfig["healthCheck"] = map[string]interface{}{
			"path":     service.LoadBalancer.HealthCheck.Path,
			"interval": service.LoadBalancer.HealthCheck.Interval,
		}
	}
	return map[string]interface{}{
		"loadBalancer": lbConfig,
	}
}

func convertMiddlewareConfigToMap(middleware *MiddlewareConfig) map[string]interface{} {
	if middleware.RedirectScheme != nil {
		return map[string]interface{}{
			"redirectScheme": map[string]interface{}{
				"scheme":    middleware.RedirectScheme.Scheme,
				"port":      middleware.RedirectScheme.Port,
				"permanent": middleware.RedirectScheme.Permanent,
			},
		}
	}
	if middleware.Headers != nil {
		result := map[string]interface{}{}
		if len(middleware.Headers.CustomRequestHeaders) > 0 {
			result["customRequestHeaders"] = middleware.Headers.CustomRequestHeaders
		}
		if len(middleware.Headers.CustomResponseHeaders) > 0 {
			result["customResponseHeaders"] = middleware.Headers.CustomResponseHeaders
		}
		if middleware.Headers.SSLRedirect {
			result["sslRedirect"] = true
		}
		return map[string]interface{}{"headers": result}
	}
	if middleware.StripPrefix != nil {
		return map[string]interface{}{
			"stripPrefix": map[string]interface{}{
				"prefixes":   middleware.StripPrefix.Prefixes,
				"forceSlash": middleware.StripPrefix.ForceSlash,
			},
		}
	}
	if middleware.AddPrefix != nil {
		return map[string]interface{}{
			"addPrefix": map[string]interface{}{
				"prefix": middleware.AddPrefix.Prefix,
			},
		}
	}
	return nil
}

func convertMapToRouterConfig(routerMap map[string]interface{}) *RouterConfig {
	config := &RouterConfig{}
	if eps, ok := routerMap["entryPoints"].([]interface{}); ok {
		config.EntryPoints = make([]string, len(eps))
		for i, ep := range eps {
			config.EntryPoints[i] = fmt.Sprintf("%v", ep)
		}
	}
	if rule, ok := routerMap["rule"].(string); ok {
		config.Rule = rule
	}
	if service, ok := routerMap["service"].(string); ok {
		config.Service = service
	}
	if middlewares, ok := routerMap["middlewares"].([]interface{}); ok {
		config.Middlewares = make([]string, len(middlewares))
		for i, m := range middlewares {
			config.Middlewares[i] = fmt.Sprintf("%v", m)
		}
	}
	if tls, ok := routerMap["tls"].(map[string]interface{}); ok {
		config.TLS = &TLSConfig{}
		if cr, ok := tls["certResolver"].(string); ok {
			config.TLS.CertResolver = cr
		}
	}
	// Handle priority - can be int or float64 from JSON unmarshaling
	if priority, ok := routerMap["priority"].(int); ok {
		config.Priority = priority
	} else if priority, ok := routerMap["priority"].(float64); ok {
		config.Priority = int(priority)
	}
	return config
}

func convertMapToServiceConfig(serviceMap map[string]interface{}) *ServiceConfig {
	config := &ServiceConfig{}
	if lb, ok := serviceMap["loadBalancer"].(map[string]interface{}); ok {
		config.LoadBalancer = &LoadBalancerConfig{}
		if servers, ok := lb["servers"].([]interface{}); ok {
			config.LoadBalancer.Servers = make([]ServerConfig, len(servers))
			for i, s := range servers {
				if serverMap, ok := s.(map[string]interface{}); ok {
					if url, ok := serverMap["url"].(string); ok {
						config.LoadBalancer.Servers[i] = ServerConfig{URL: url}
					}
				}
			}
		}
		if phh, ok := lb["passHostHeader"].(bool); ok {
			config.LoadBalancer.PassHostHeader = phh
		}
		if hc, ok := lb["healthCheck"].(map[string]interface{}); ok {
			config.LoadBalancer.HealthCheck = &HealthCheck{}
			if path, ok := hc["path"].(string); ok {
				config.LoadBalancer.HealthCheck.Path = path
			}
			if interval, ok := hc["interval"].(string); ok {
				config.LoadBalancer.HealthCheck.Interval = interval
			}
		}
	}
	return config
}

func convertMapToMiddlewareConfig(middlewareMap map[string]interface{}) *MiddlewareConfig {
	config := &MiddlewareConfig{}

	if rs, ok := middlewareMap["redirectScheme"].(map[string]interface{}); ok {
		config.RedirectScheme = &RedirectSchemeConfig{}
		if scheme, ok := rs["scheme"].(string); ok {
			config.RedirectScheme.Scheme = scheme
		}
		if port, ok := rs["port"].(string); ok {
			config.RedirectScheme.Port = port
		}
		if permanent, ok := rs["permanent"].(bool); ok {
			config.RedirectScheme.Permanent = permanent
		}
	}

	if headers, ok := middlewareMap["headers"].(map[string]interface{}); ok {
		config.Headers = &HeadersConfig{}
		if crh, ok := headers["customRequestHeaders"].(map[string]interface{}); ok {
			config.Headers.CustomRequestHeaders = make(map[string]string)
			for k, v := range crh {
				config.Headers.CustomRequestHeaders[k] = fmt.Sprintf("%v", v)
			}
		}
		if crh, ok := headers["customResponseHeaders"].(map[string]interface{}); ok {
			config.Headers.CustomResponseHeaders = make(map[string]string)
			for k, v := range crh {
				config.Headers.CustomResponseHeaders[k] = fmt.Sprintf("%v", v)
			}
		}
		if sr, ok := headers["sslRedirect"].(bool); ok {
			config.Headers.SSLRedirect = sr
		}
	}

	if sp, ok := middlewareMap["stripPrefix"].(map[string]interface{}); ok {
		config.StripPrefix = &StripPrefixConfig{}
		if prefixes, ok := sp["prefixes"].([]interface{}); ok {
			config.StripPrefix.Prefixes = make([]string, len(prefixes))
			for i, p := range prefixes {
				config.StripPrefix.Prefixes[i] = fmt.Sprintf("%v", p)
			}
		}
		if fs, ok := sp["forceSlash"].(bool); ok {
			config.StripPrefix.ForceSlash = fs
		}
	}

	if ap, ok := middlewareMap["addPrefix"].(map[string]interface{}); ok {
		config.AddPrefix = &AddPrefixConfig{}
		if prefix, ok := ap["prefix"].(string); ok {
			config.AddPrefix.Prefix = prefix
		}
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
