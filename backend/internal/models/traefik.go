package models

import (
	"time"
)

// Router represents a Traefik router configuration
type Router struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	Name      string           `gorm:"uniqueIndex;not null" json:"name"` // Unique router name
	Hostnames []RouterHostname `gorm:"foreignKey:RouterID;constraint:OnDelete:CASCADE" json:"hostnames"`
	ServiceID uint             `gorm:"not null" json:"service_id"`
	Service   Service          `gorm:"foreignKey:ServiceID" json:"service,omitempty"`

	// Ownership - who created this proxy
	UserID uint `gorm:"not null;index" json:"user_id"`

	// TLS Configuration
	TLSEnabled      bool   `gorm:"default:false" json:"tls_enabled"`             // Enable TLS
	TLSCertResolver string `gorm:"default:letsencrypt" json:"tls_cert_resolver"` // ACME resolver name

	// Redirect to HTTPS
	RedirectHTTPS bool `gorm:"default:true" json:"redirect_https"` // Redirect HTTP to HTTPS

	// EntryPoints (comma-separated or stored in separate table)
	EntryPoints string `gorm:"default:web,websecure" json:"entry_points"` // web, websecure

	// Middleware associations
	Middlewares []RouterMiddleware `gorm:"foreignKey:RouterID" json:"middlewares,omitempty"`

	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RouterHostname represents hostnames/domains for a router
type RouterHostname struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	RouterID uint   `gorm:"not null;index" json:"router_id"`
	Hostname string `gorm:"not null" json:"hostname"` // e.g., app.example.com
}

// RouterMiddleware links routers to middlewares (many-to-many)
type RouterMiddleware struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	RouterID     uint       `gorm:"not null;index" json:"router_id"`
	MiddlewareID uint       `gorm:"not null;index" json:"middleware_id"`
	Middleware   Middleware `gorm:"foreignKey:MiddlewareID" json:"middleware,omitempty"`
	Priority     int        `gorm:"default:0" json:"priority"` // Execution order
}

// Service represents a Traefik service configuration
type Service struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"` // Unique service name
	Type string `gorm:"default:http" json:"type"`         // http (for now)

	// Load Balancing
	Servers          []ServiceServer `gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE" json:"servers"`
	LoadBalancerType string          `gorm:"default:wrr" json:"load_balancer_type"` // wrr (weighted round robin)

	// Pass Host Header
	PassHostHeader bool `gorm:"default:true" json:"pass_host_header"`

	// Health Check (basic support)
	HealthCheckEnabled  bool   `gorm:"default:false" json:"health_check_enabled"`
	HealthCheckPath     string `json:"health_check_path,omitempty"`
	HealthCheckInterval int    `gorm:"default:10" json:"health_check_interval"` // seconds

	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServiceServer represents a backend server for load balancing
type ServiceServer struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ServiceID uint   `gorm:"not null;index" json:"service_id"`
	URL       string `gorm:"not null" json:"url"` // e.g., http://192.168.1.100:8080
	Weight    int    `gorm:"default:1" json:"weight"`
	IsHealthy *bool  `json:"is_healthy,omitempty"` // Traefik tracks this
}

// Middleware represents a Traefik middleware configuration
type Middleware struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"` // Unique middleware name
	Type string `gorm:"not null" json:"type"`             // redirectScheme, headers, stripPrefix, etc.

	// Type-specific configuration stored as JSON
	Config string `gorm:"type:text" json:"config"` // JSON configuration based on type

	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Request/Response structures for API

type CreateRouterRequest struct {
	Name            string   `json:"name" binding:"required"`
	Hostnames       []string `json:"hostnames" binding:"required,min=1"`
	ServiceID       uint     `json:"service_id" binding:"required"`
	TLSEnabled      bool     `json:"tls_enabled"`
	TLSCertResolver string   `json:"tls_cert_resolver,omitempty"`
	RedirectHTTPS   bool     `json:"redirect_https"`
	EntryPoints     []string `json:"entry_points,omitempty"`
	MiddlewareIDs   []uint   `json:"middleware_ids,omitempty"`
}

type UpdateRouterRequest struct {
	Name            *string  `json:"name,omitempty"`
	Hostnames       []string `json:"hostnames,omitempty"`
	ServiceID       *uint    `json:"service_id,omitempty"`
	TLSEnabled      *bool    `json:"tls_enabled,omitempty"`
	TLSCertResolver *string  `json:"tls_cert_resolver,omitempty"`
	RedirectHTTPS   *bool    `json:"redirect_https,omitempty"`
	EntryPoints     []string `json:"entry_points,omitempty"`
	MiddlewareIDs   []uint   `json:"middleware_ids,omitempty"`
	IsActive        *bool    `json:"is_active,omitempty"`
}

type CreateServiceRequest struct {
	Name                string   `json:"name" binding:"required"`
	Servers             []string `json:"servers" binding:"required,min=1"` // URLs
	LoadBalancerType    string   `json:"load_balancer_type,omitempty" binding:"omitempty,oneof=wrr drr"`
	PassHostHeader      bool     `json:"pass_host_header"`
	HealthCheckEnabled  bool     `json:"health_check_enabled"`
	HealthCheckPath     string   `json:"health_check_path,omitempty"`
	HealthCheckInterval int      `json:"health_check_interval,omitempty"`
}

type UpdateServiceRequest struct {
	Name                *string  `json:"name,omitempty"`
	Servers             []string `json:"servers,omitempty"`
	LoadBalancerType    *string  `json:"load_balancer_type,omitempty" binding:"omitempty,oneof=wrr drr"`
	PassHostHeader      *bool    `json:"pass_host_header,omitempty"`
	HealthCheckEnabled  *bool    `json:"health_check_enabled,omitempty"`
	HealthCheckPath     *string  `json:"health_check_path,omitempty"`
	HealthCheckInterval *int     `json:"health_check_interval,omitempty"`
	IsActive            *bool    `json:"is_active,omitempty"`
}

type CreateMiddlewareRequest struct {
	Name   string           `json:"name" binding:"required"`
	Type   string           `json:"type" binding:"required,oneof=redirectScheme headers stripPrefix addPrefix"`
	Config MiddlewareConfig `json:"config" binding:"required"`
}

type UpdateMiddlewareRequest struct {
	Name     *string           `json:"name,omitempty"`
	Type     *string           `json:"type,omitempty" binding:"omitempty,oneof=redirectScheme headers stripPrefix addPrefix"`
	Config   *MiddlewareConfig `json:"config,omitempty"`
	IsActive *bool             `json:"is_active,omitempty"`
}

// MiddlewareConfig for different middleware types
type MiddlewareConfig struct {
	// RedirectScheme
	Scheme    string `json:"scheme,omitempty"`    // https
	Permanent bool   `json:"permanent,omitempty"` // true for 301, false for 302
	Port      string `json:"port,omitempty"`      // 443

	// Headers
	CustomRequestHeaders  map[string]string `json:"customRequestHeaders,omitempty"`
	CustomResponseHeaders map[string]string `json:"customResponseHeaders,omitempty"`
	SSLRedirect           bool              `json:"sslRedirect,omitempty"`

	// StripPrefix
	Prefixes   []string `json:"prefixes,omitempty"`
	ForceSlash bool     `json:"forceSlash,omitempty"`

	// AddPrefix
	Prefix string `json:"prefix,omitempty"`
}

type RouterResponse struct {
	ID              uint             `json:"id"`
	Name            string           `json:"name"`
	Hostnames       []string         `json:"hostnames"`
	ServiceID       uint             `json:"service_id"`
	ServiceName     string           `json:"service_name"`
	TLSEnabled      bool             `json:"tls_enabled"`
	TLSCertResolver string           `json:"tls_cert_resolver"`
	RedirectHTTPS   bool             `json:"redirect_https"`
	EntryPoints     []string         `json:"entry_points"`
	Middlewares     []MiddlewareInfo `json:"middlewares"`
	IsActive        bool             `json:"is_active"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type MiddlewareInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Priority int    `json:"priority"`
}

// ToResponse converts Router to RouterResponse
func (r *Router) ToResponse() RouterResponse {
	hostnames := make([]string, len(r.Hostnames))
	for i, h := range r.Hostnames {
		hostnames[i] = h.Hostname
	}

	middlewares := make([]MiddlewareInfo, len(r.Middlewares))
	for i, m := range r.Middlewares {
		middlewares[i] = MiddlewareInfo{
			ID:       m.MiddlewareID,
			Name:     m.Middleware.Name,
			Type:     m.Middleware.Type,
			Priority: m.Priority,
		}
	}

	serviceName := ""
	if r.Service.ID > 0 {
		serviceName = r.Service.Name
	}

	return RouterResponse{
		ID:              r.ID,
		Name:            r.Name,
		Hostnames:       hostnames,
		ServiceID:       r.ServiceID,
		ServiceName:     serviceName,
		TLSEnabled:      r.TLSEnabled,
		TLSCertResolver: r.TLSCertResolver,
		RedirectHTTPS:   r.RedirectHTTPS,
		EntryPoints:     splitEntryPoints(r.EntryPoints),
		Middlewares:     middlewares,
		IsActive:        r.IsActive,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// Helper to split comma-separated entry points
func splitEntryPoints(ep string) []string {
	if ep == "" {
		return []string{"web", "websecure"}
	}
	return splitAndTrim(ep)
}

func splitAndTrim(s string) []string {
	parts := []string{}
	for _, p := range split(s, ",") {
		if trimmed := trimSpace(p); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s, sep string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if i < len(s)-len(sep)+1 && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

type ServiceResponse struct {
	ID                  uint             `json:"id"`
	Name                string           `json:"name"`
	Type                string           `json:"type"`
	Servers             []ServerResponse `json:"servers"`
	LoadBalancerType    string           `json:"load_balancer_type"`
	PassHostHeader      bool             `json:"pass_host_header"`
	HealthCheckEnabled  bool             `json:"health_check_enabled"`
	HealthCheckPath     string           `json:"health_check_path,omitempty"`
	HealthCheckInterval int              `json:"health_check_interval,omitempty"`
	IsActive            bool             `json:"is_active"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type ServerResponse struct {
	ID     uint   `json:"id"`
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

// ToResponse converts Service to ServiceResponse
func (s *Service) ToResponse() ServiceResponse {
	servers := make([]ServerResponse, len(s.Servers))
	for i, srv := range s.Servers {
		servers[i] = ServerResponse{
			ID:     srv.ID,
			URL:    srv.URL,
			Weight: srv.Weight,
		}
	}

	return ServiceResponse{
		ID:                  s.ID,
		Name:                s.Name,
		Type:                s.Type,
		Servers:             servers,
		LoadBalancerType:    s.LoadBalancerType,
		PassHostHeader:      s.PassHostHeader,
		HealthCheckEnabled:  s.HealthCheckEnabled,
		HealthCheckPath:     s.HealthCheckPath,
		HealthCheckInterval: s.HealthCheckInterval,
		IsActive:            s.IsActive,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}

type MiddlewareResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Config    string    `json:"config,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Middleware to MiddlewareResponse
func (m *Middleware) ToResponse() MiddlewareResponse {
	return MiddlewareResponse{
		ID:        m.ID,
		Name:      m.Name,
		Type:      m.Type,
		Config:    m.Config,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// HTTPProvider represents an external Traefik HTTP Provider to aggregate
type HTTPProvider struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"uniqueIndex;not null" json:"name"`
	URL             string     `gorm:"not null" json:"url"`
	Priority        int        `gorm:"default:0;index" json:"priority"` // Higher = higher priority
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	RefreshInterval int        `gorm:"default:30" json:"refresh_interval"` // seconds
	LastFetched     *time.Time `json:"last_fetched"`
	LastResponse    []byte     `gorm:"type:text" json:"-"`
	LastError       string     `json:"last_error"`
	RouterCount     int        `json:"router_count"`
	ServiceCount    int        `json:"service_count"`
	MiddlewareCount int        `json:"middleware_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToResponse converts HTTPProvider to a safe response
func (e *HTTPProvider) ToResponse() map[string]interface{} {
	lastFetched := ""
	if e.LastFetched != nil {
		lastFetched = e.LastFetched.Format(time.RFC3339)
	}
	return map[string]interface{}{
		"id":               e.ID,
		"name":             e.Name,
		"url":              e.URL,
		"priority":         e.Priority,
		"is_active":        e.IsActive,
		"refresh_interval": e.RefreshInterval,
		"last_fetched":     lastFetched,
		"last_error":       e.LastError,
		"router_count":     e.RouterCount,
		"service_count":    e.ServiceCount,
		"middleware_count": e.MiddlewareCount,
		"created_at":       e.CreatedAt.Format(time.RFC3339),
		"updated_at":       e.UpdatedAt.Format(time.RFC3339),
	}
}

// Request/Response structures for HTTPProvider API

type CreateHTTPProviderRequest struct {
	Name            string `json:"name" binding:"required"`
	URL             string `json:"url" binding:"required,url"`
	Priority        int    `json:"priority"`
	RefreshInterval int    `json:"refresh_interval"`
	IsActive        bool   `json:"is_active"`
}

type UpdateHTTPProviderRequest struct {
	Name            *string `json:"name,omitempty"`
	URL             *string `json:"url,omitempty" binding:"omitempty,url"`
	Priority        *int    `json:"priority,omitempty"`
	RefreshInterval *int    `json:"refresh_interval,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
}
