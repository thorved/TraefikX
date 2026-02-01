package models

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/types"
)

// TraefikDynamicConfig represents the full Traefik dynamic configuration
// This uses the official Traefik v3 types for maximum compatibility
type TraefikDynamicConfig struct {
	HTTP *dynamic.HTTPConfiguration `json:"http,omitempty"`
	TCP  *dynamic.TCPConfiguration  `json:"tcp,omitempty"`
	TLS  *dynamic.TLSConfiguration  `json:"tls,omitempty"`
}

// TraefikRouter is an alias for the official Traefik Router type
type TraefikRouter = dynamic.Router

// TraefikRouterTLSConfig is an alias for the official Traefik RouterTLSConfig type
type TraefikRouterTLSConfig = dynamic.RouterTLSConfig

// TraefikService is an alias for the official Traefik Service type
type TraefikService = dynamic.Service

// TraefikServersLoadBalancer is an alias for the official Traefik ServersLoadBalancer type
type TraefikServersLoadBalancer = dynamic.ServersLoadBalancer

// TraefikServer is an alias for the official Traefik Server type
type TraefikServer = dynamic.Server

// TraefikServerHealthCheck is an alias for the official Traefik ServerHealthCheck type
type TraefikServerHealthCheck = dynamic.ServerHealthCheck

// TraefikMiddleware is an alias for the official Traefik Middleware type
type TraefikMiddleware = dynamic.Middleware

// TraefikServersTransport is an alias for the official Traefik ServersTransport type
type TraefikServersTransport = dynamic.ServersTransport

// TraefikForwardingTimeouts is an alias for the official Traefik ForwardingTimeouts type
type TraefikForwardingTimeouts = dynamic.ForwardingTimeouts

// TraefikDomain is an alias for the official Traefik Domain type
type TraefikDomain = types.Domain

// TraefikHTTPConfiguration is an alias for the official Traefik HTTPConfiguration type
type TraefikHTTPConfiguration = dynamic.HTTPConfiguration

// TraefikTCPConfiguration is an alias for the official Traefik TCPConfiguration type
type TraefikTCPConfiguration = dynamic.TCPConfiguration

// TraefikTCPRouter is an alias for the official Traefik TCPRouter type
type TraefikTCPRouter = dynamic.TCPRouter

// TraefikTCPService is an alias for the official Traefik TCPService type
type TraefikTCPService = dynamic.TCPService

// TraefikTCPServersLoadBalancer is an alias for the official Traefik TCPServersLoadBalancer type
type TraefikTCPServersLoadBalancer = dynamic.TCPServersLoadBalancer

// TraefikTCPServer is an alias for the official Traefik TCPServer type
type TraefikTCPServer = dynamic.TCPServer

// TraefikTCPServersTransport is an alias for the official Traefik TCPServersTransport type
type TraefikTCPServersTransport = dynamic.TCPServersTransport

// TraefikTLSClientConfig is an alias for the official Traefik TLSClientConfig type
type TraefikTLSClientConfig = dynamic.TLSClientConfig

// Common middleware configuration types (for convenient access)

// TraefikRedirectScheme is an alias for the official Traefik RedirectScheme type
type TraefikRedirectScheme = dynamic.RedirectScheme

// TraefikHeaders is an alias for the official Traefik Headers type
type TraefikHeaders = dynamic.Headers

// TraefikStripPrefix is an alias for the official Traefik StripPrefix type
type TraefikStripPrefix = dynamic.StripPrefix

// TraefikAddPrefix is an alias for the official Traefik AddPrefix type
type TraefikAddPrefix = dynamic.AddPrefix

// TraefikReplacePath is an alias for the official Traefik ReplacePath type
type TraefikReplacePath = dynamic.ReplacePath

// TraefikReplacePathRegex is an alias for the official Traefik ReplacePathRegex type
type TraefikReplacePathRegex = dynamic.ReplacePathRegex

// TraefikChain is an alias for the official Traefik Chain type
type TraefikChain = dynamic.Chain

// TraefikIPAllowList is an alias for the official Traefik IPAllowList type
type TraefikIPAllowList = dynamic.IPAllowList

// TraefikIPStrategy is an alias for the official Traefik IPStrategy type
type TraefikIPStrategy = dynamic.IPStrategy

// TraefikRateLimit is an alias for the official Traefik RateLimit type
type TraefikRateLimit = dynamic.RateLimit

// TraefikCompress is an alias for the official Traefik Compress type
type TraefikCompress = dynamic.Compress

// TraefikBasicAuth is an alias for the official Traefik BasicAuth type
type TraefikBasicAuth = dynamic.BasicAuth

// TraefikForwardAuth is an alias for the official Traefik ForwardAuth type
type TraefikForwardAuth = dynamic.ForwardAuth

// TraefikBuffering is an alias for the official Traefik Buffering type
type TraefikBuffering = dynamic.Buffering

// TraefikCircuitBreaker is an alias for the official Traefik CircuitBreaker type
type TraefikCircuitBreaker = dynamic.CircuitBreaker

// TraefikRetry is an alias for the official Traefik Retry type
type TraefikRetry = dynamic.Retry

// TraefikWeightedRoundRobin is an alias for the official Traefik WeightedRoundRobin type
type TraefikWeightedRoundRobin = dynamic.WeightedRoundRobin

// TraefikWRRService is an alias for the official Traefik WRRService type
type TraefikWRRService = dynamic.WRRService

// TraefikMirroring is an alias for the official Traefik Mirroring type
type TraefikMirroring = dynamic.Mirroring

// TraefikFailover is an alias for the official Traefik Failover type
type TraefikFailover = dynamic.Failover

// TraefikResponseForwarding is an alias for the official Traefik ResponseForwarding type
type TraefikResponseForwarding = dynamic.ResponseForwarding

// TraefikPassiveServerHealthCheck is an alias for the official Traefik PassiveServerHealthCheck type
type TraefikPassiveServerHealthCheck = dynamic.PassiveServerHealthCheck

// TraefikHealthCheck is an alias for the official Traefik HealthCheck type (for services)
type TraefikHealthCheck = dynamic.HealthCheck

// TraefikSticky is an alias for the official Traefik Sticky type
type TraefikSticky = dynamic.Sticky

// TraefikCookie is an alias for the official Traefik Cookie type
type TraefikCookie = dynamic.Cookie

// NewTraefikDynamicConfig creates a new empty dynamic configuration
func NewTraefikDynamicConfig() *TraefikDynamicConfig {
	return &TraefikDynamicConfig{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Services:          make(map[string]*dynamic.Service),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Models:            make(map[string]*dynamic.Model),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			Middlewares:       make(map[string]*dynamic.TCPMiddleware),
			Models:            make(map[string]*dynamic.TCPModel),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
		},
	}
}

// NewTraefikHTTPConfiguration creates a new HTTP configuration with initialized maps
func NewTraefikHTTPConfiguration() *dynamic.HTTPConfiguration {
	return &dynamic.HTTPConfiguration{
		Routers:           make(map[string]*dynamic.Router),
		Services:          make(map[string]*dynamic.Service),
		Middlewares:       make(map[string]*dynamic.Middleware),
		Models:            make(map[string]*dynamic.Model),
		ServersTransports: make(map[string]*dynamic.ServersTransport),
	}
}

// NewTraefikTCPConfiguration creates a new TCP configuration with initialized maps
func NewTraefikTCPConfiguration() *dynamic.TCPConfiguration {
	return &dynamic.TCPConfiguration{
		Routers:           make(map[string]*dynamic.TCPRouter),
		Services:          make(map[string]*dynamic.TCPService),
		Middlewares:       make(map[string]*dynamic.TCPMiddleware),
		Models:            make(map[string]*dynamic.TCPModel),
		ServersTransports: make(map[string]*dynamic.TCPServersTransport),
	}
}
