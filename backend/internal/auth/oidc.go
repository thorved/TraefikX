package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/traefikx/backend/internal/config"
	"golang.org/x/oauth2"
)

var (
	oauthConfig    *oauth2.Config
	oidcStateStore = make(map[string]*OIDCState)
)

type OIDCState struct {
	State      string
	ExpiresAt  time.Time
	LinkToUser uint // If > 0, link to existing user instead of creating new
}

// InitOIDC initializes the OIDC configuration
func InitOIDC(cfg *config.Config) error {
	if !cfg.OIDCEnabled {
		return nil
	}

	if cfg.OIDCIssuerURL == "" || cfg.OIDCClientID == "" || cfg.OIDCClientSecret == "" {
		return errors.New("OIDC configuration incomplete: issuer URL, client ID, and client secret are required")
	}

	// Use configured endpoints or fall back to issuer-based discovery (simplified for now to require config)
	authURL := cfg.OIDCAuthURL
	if authURL == "" {
		authURL = cfg.OIDCIssuerURL + "/oauth/authorize"
	}

	tokenURL := cfg.OIDCTokenURL
	if tokenURL == "" {
		tokenURL = cfg.OIDCIssuerURL + "/oauth/token"
	}

	oauthConfig = &oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURL,
		Scopes:       cfg.OIDCScopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	// Clean up expired states periodically
	go cleanupExpiredStates()

	return nil
}

// GenerateOIDCState generates a random state for OIDC flow
func GenerateOIDCState(linkToUser uint) string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	oidcStateStore[state] = &OIDCState{
		State:      state,
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		LinkToUser: linkToUser,
	}

	return state
}

// ValidateOIDCState validates and returns the state
func ValidateOIDCState(state string) (*OIDCState, bool) {
	oidcState, exists := oidcStateStore[state]
	if !exists {
		return nil, false
	}

	if time.Now().After(oidcState.ExpiresAt) {
		delete(oidcStateStore, state)
		return nil, false
	}

	// Clean up after use
	delete(oidcStateStore, state)
	return oidcState, true
}

// GetOIDCAuthURL returns the authorization URL for OIDC login
func GetOIDCAuthURL(state string) string {
	if oauthConfig == nil {
		return ""
	}
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// ExchangeOIDCCode exchanges authorization code for tokens
func ExchangeOIDCCode(code string) (*oauth2.Token, error) {
	if oauthConfig == nil {
		return nil, errors.New("OIDC not configured")
	}

	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// GetOIDCUserInfo extracts user info from OIDC token
// This is a simplified version - in production, you'd verify the ID token
func GetOIDCUserInfo(token *oauth2.Token) (*OIDCUserInfo, error) {
	if oauthConfig == nil {
		return nil, errors.New("OIDC not configured")
	}

	cfg := config.AppConfig

	// Fetch user info from userinfo endpoint
	// Fetch user info from userinfo endpoint
	userInfoURL := cfg.OIDCUserInfoURL
	if userInfoURL == "" {
		userInfoURL = cfg.OIDCIssuerURL + "/oidc/userinfo"
	}

	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}

	// Parse user info
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo OIDCUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}

type OIDCUserInfo struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

func cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		now := time.Now()
		for state, oidcState := range oidcStateStore {
			if now.After(oidcState.ExpiresAt) {
				delete(oidcStateStore, state)
			}
		}
	}
}

// IsOIDCEnabled returns whether OIDC is configured and enabled
func IsOIDCEnabled() bool {
	return oauthConfig != nil
}
