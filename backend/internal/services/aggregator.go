package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/traefikx/backend/internal/models"
	"gorm.io/gorm"
)

// TraefikHTTPConfig represents the Traefik HTTP configuration structure
type TraefikHTTPConfig struct {
	Routers     map[string]interface{} `json:"routers,omitempty"`
	Services    map[string]interface{} `json:"services,omitempty"`
	Middlewares map[string]interface{} `json:"middlewares,omitempty"`
}

// TraefikConfig represents the full Traefik configuration
type TraefikConfig struct {
	HTTP TraefikHTTPConfig `json:"http,omitempty"`
}

// ProviderStatus represents the current status of a provider
type ProviderStatus struct {
	ID              uint
	Name            string
	URL             string
	Priority        int
	IsActive        bool
	LastFetched     *time.Time
	LastError       string
	Config          *TraefikConfig
	RouterCount     int
	ServiceCount    int
	MiddlewareCount int
}

// AggregatorService manages polling and caching of external HTTP providers
type AggregatorService struct {
	db         *gorm.DB
	client     *http.Client
	statuses   map[uint]*ProviderStatus
	statusesMu sync.RWMutex
	tickers    map[uint]*time.Ticker
	tickersMu  sync.Mutex
	stopChan   chan struct{}
}

// NewAggregatorService creates a new aggregator service
func NewAggregatorService(db *gorm.DB) *AggregatorService {
	return &AggregatorService{
		db:       db,
		client:   &http.Client{Timeout: 5 * time.Second},
		statuses: make(map[uint]*ProviderStatus),
		tickers:  make(map[uint]*time.Ticker),
		stopChan: make(chan struct{}),
	}
}

// Start begins the aggregator service
func (a *AggregatorService) Start() {
	log.Println("Starting HTTP provider aggregator service...")

	// Load all providers and start polling
	var providers []models.HTTPProvider
	if err := a.db.Find(&providers).Error; err != nil {
		log.Printf("Failed to load HTTP providers: %v", err)
		return
	}

	for _, provider := range providers {
		a.startPolling(&provider)
	}

	log.Printf("Aggregator service started with %d providers", len(providers))
}

// Stop stops all polling
func (a *AggregatorService) Stop() {
	close(a.stopChan)
	a.tickersMu.Lock()
	for _, ticker := range a.tickers {
		ticker.Stop()
	}
	a.tickers = make(map[uint]*time.Ticker)
	a.tickersMu.Unlock()
}

// startPolling starts polling for a specific provider
func (a *AggregatorService) startPolling(provider *models.HTTPProvider) {
	a.tickersMu.Lock()
	defer a.tickersMu.Unlock()

	// Stop existing ticker if any
	if ticker, exists := a.tickers[provider.ID]; exists {
		ticker.Stop()
	}

	if !provider.IsActive {
		return
	}

	// Do initial fetch
	a.fetchProvider(provider)

	// Start ticker
	interval := time.Duration(provider.RefreshInterval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second // Minimum 5 seconds
	}

	ticker := time.NewTicker(interval)
	a.tickers[provider.ID] = ticker

	go func(providerID uint) {
		for {
			select {
			case <-ticker.C:
				var p models.HTTPProvider
				if err := a.db.First(&p, providerID).Error; err != nil {
					log.Printf("Provider %d not found, stopping polling", providerID)
					return
				}
				a.fetchProvider(&p)
			case <-a.stopChan:
				return
			}
		}
	}(provider.ID)
}

// fetchProvider fetches configuration from a provider
func (a *AggregatorService) fetchProvider(provider *models.HTTPProvider) {
	log.Printf("Fetching from provider %s (%s)", provider.Name, provider.URL)

	resp, err := a.client.Get(provider.URL)
	if err != nil {
		a.updateProviderError(provider, fmt.Sprintf("Connection error: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.updateProviderError(provider, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.updateProviderError(provider, fmt.Sprintf("Read error: %v", err))
		return
	}

	var config TraefikConfig
	if err := json.Unmarshal(body, &config); err != nil {
		a.updateProviderError(provider, fmt.Sprintf("JSON parse error: %v", err))
		return
	}

	// Count items
	routerCount := len(config.HTTP.Routers)
	serviceCount := len(config.HTTP.Services)
	middlewareCount := len(config.HTTP.Middlewares)

	// Update database
	now := time.Now()
	provider.LastFetched = &now
	provider.LastResponse = body
	provider.LastError = ""
	provider.RouterCount = routerCount
	provider.ServiceCount = serviceCount
	provider.MiddlewareCount = middlewareCount

	if err := a.db.Save(provider).Error; err != nil {
		log.Printf("Failed to save provider %s: %v", provider.Name, err)
	}

	// Update in-memory status
	a.statusesMu.Lock()
	a.statuses[provider.ID] = &ProviderStatus{
		ID:              provider.ID,
		Name:            provider.Name,
		URL:             provider.URL,
		Priority:        provider.Priority,
		IsActive:        provider.IsActive,
		LastFetched:     provider.LastFetched,
		LastError:       "",
		Config:          &config,
		RouterCount:     routerCount,
		ServiceCount:    serviceCount,
		MiddlewareCount: middlewareCount,
	}
	a.statusesMu.Unlock()

	log.Printf("Successfully fetched from %s: %d routers, %d services, %d middlewares",
		provider.Name, routerCount, serviceCount, middlewareCount)
}

// updateProviderError updates provider with error status
func (a *AggregatorService) updateProviderError(provider *models.HTTPProvider, errMsg string) {
	log.Printf("Provider %s error: %s", provider.Name, errMsg)

	provider.LastError = errMsg
	if err := a.db.Model(provider).Update("last_error", errMsg).Error; err != nil {
		log.Printf("Failed to update provider error: %v", err)
	}

	a.statusesMu.Lock()
	if status, exists := a.statuses[provider.ID]; exists {
		status.LastError = errMsg
	}
	a.statusesMu.Unlock()
}

// RefreshProvider manually refreshes a specific provider
func (a *AggregatorService) RefreshProvider(providerID uint) error {
	var provider models.HTTPProvider
	if err := a.db.First(&provider, providerID).Error; err != nil {
		return err
	}

	go a.fetchProvider(&provider)
	return nil
}

// AddProvider adds a new provider and starts polling
func (a *AggregatorService) AddProvider(provider *models.HTTPProvider) {
	a.startPolling(provider)
}

// UpdateProvider updates a provider and restarts polling
func (a *AggregatorService) UpdateProvider(provider *models.HTTPProvider) {
	a.startPolling(provider)
}

// DeleteProvider stops polling for a provider
func (a *AggregatorService) DeleteProvider(providerID uint) {
	a.tickersMu.Lock()
	if ticker, exists := a.tickers[providerID]; exists {
		ticker.Stop()
		delete(a.tickers, providerID)
	}
	a.tickersMu.Unlock()

	a.statusesMu.Lock()
	delete(a.statuses, providerID)
	a.statusesMu.Unlock()
}

// GetStatuses returns all provider statuses
func (a *AggregatorService) GetStatuses() []ProviderStatus {
	a.statusesMu.RLock()
	defer a.statusesMu.RUnlock()

	statuses := make([]ProviderStatus, 0, len(a.statuses))
	for _, status := range a.statuses {
		statuses = append(statuses, *status)
	}

	// Sort by priority (higher first)
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Priority > statuses[j].Priority
	})

	return statuses
}

// GetMergedConfig returns the merged configuration from all active providers
// Priority: Local DB > Provider (by priority, higher first)
func (a *AggregatorService) GetMergedConfig(localRouters, localServices, localMiddlewares map[string]interface{}) (*TraefikConfig, []ConflictInfo) {
	a.statusesMu.RLock()
	defer a.statusesMu.RUnlock()

	merged := &TraefikConfig{
		HTTP: TraefikHTTPConfig{
			Routers:     make(map[string]interface{}),
			Services:    make(map[string]interface{}),
			Middlewares: make(map[string]interface{}),
		},
	}
	conflicts := []ConflictInfo{}

	// Track which items came from which source for conflict detection
	routerSources := make(map[string]string)
	serviceSources := make(map[string]string)
	middlewareSources := make(map[string]string)

	// Add local config first (highest priority)
	for name, router := range localRouters {
		merged.HTTP.Routers[name] = router
		routerSources[name] = "local"
	}
	for name, service := range localServices {
		merged.HTTP.Services[name] = service
		serviceSources[name] = "local"
	}
	for name, middleware := range localMiddlewares {
		merged.HTTP.Middlewares[name] = middleware
		middlewareSources[name] = "local"
	}

	// Get sorted statuses by priority (higher first)
	statuses := make([]*ProviderStatus, 0, len(a.statuses))
	for _, status := range a.statuses {
		if status.IsActive && status.Config != nil && status.LastError == "" {
			statuses = append(statuses, status)
		}
	}
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Priority > statuses[j].Priority
	})

	// Merge from providers
	for _, status := range statuses {
		if status.Config == nil {
			continue
		}

		// Merge routers
		for name, router := range status.Config.HTTP.Routers {
			if existingSource, exists := routerSources[name]; exists {
				if existingSource != status.Name {
					conflicts = append(conflicts, ConflictInfo{
						Type:           "router",
						Name:           name,
						Source:         status.Name,
						OverriddenBy:   existingSource,
						SourcePriority: status.Priority,
					})
				}
				continue // Skip, higher priority already has it
			}
			merged.HTTP.Routers[name] = router
			routerSources[name] = status.Name
		}

		// Merge services
		for name, service := range status.Config.HTTP.Services {
			if existingSource, exists := serviceSources[name]; exists {
				if existingSource != status.Name {
					conflicts = append(conflicts, ConflictInfo{
						Type:           "service",
						Name:           name,
						Source:         status.Name,
						OverriddenBy:   existingSource,
						SourcePriority: status.Priority,
					})
				}
				continue
			}
			merged.HTTP.Services[name] = service
			serviceSources[name] = status.Name
		}

		// Merge middlewares
		for name, middleware := range status.Config.HTTP.Middlewares {
			if existingSource, exists := middlewareSources[name]; exists {
				if existingSource != status.Name {
					conflicts = append(conflicts, ConflictInfo{
						Type:           "middleware",
						Name:           name,
						Source:         status.Name,
						OverriddenBy:   existingSource,
						SourcePriority: status.Priority,
					})
				}
				continue
			}
			merged.HTTP.Middlewares[name] = middleware
			middlewareSources[name] = status.Name
		}
	}

	return merged, conflicts
}

// ConflictInfo represents a configuration conflict
type ConflictInfo struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	Source         string `json:"source"`
	OverriddenBy   string `json:"overridden_by"`
	SourcePriority int    `json:"source_priority"`
}

// GetProviderResponse returns the last cached response for a provider
func (a *AggregatorService) GetProviderResponse(providerID uint) ([]byte, error) {
	var provider models.HTTPProvider
	if err := a.db.First(&provider, providerID).Error; err != nil {
		return nil, err
	}

	if len(provider.LastResponse) == 0 {
		return nil, fmt.Errorf("no cached response available")
	}

	return provider.LastResponse, nil
}
