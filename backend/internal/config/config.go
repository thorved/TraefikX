package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port      string
	JWTSecret string
	Env       string

	// Database
	DatabasePath string

	// Security
	BcryptCost           int
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration

	// OIDC
	OIDCEnabled      bool
	OIDCProviderName string
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	OIDCScopes       []string
	OIDCAuthURL      string
	OIDCTokenURL     string
	OIDCUserInfoURL  string

	// CORS
	CORSAllowedOrigins []string

	// Default Admin
	DefaultAdminEmail    string
	DefaultAdminPassword string

	// Traefik HTTP Provider
	TraefikProviderToken string // Token for /api/provider endpoint authentication
}

var AppConfig *Config

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		// Server defaults
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "change-this-secret-key-in-production-min-32-chars"),
		Env:       getEnv("ENV", "development"),

		// Database defaults
		DatabasePath: getEnv("DATABASE_PATH", "./data/traefikx.db"),

		// Security defaults
		BcryptCost:           getEnvAsInt("BCRYPT_COST", 12),
		AccessTokenDuration:  getEnvAsDuration("ACCESS_TOKEN_DURATION", 15*time.Minute),
		RefreshTokenDuration: getEnvAsDuration("REFRESH_TOKEN_DURATION", 7*24*time.Hour),

		// OIDC defaults
		OIDCEnabled:      getEnvAsBool("OIDC_ENABLED", false),
		OIDCProviderName: getEnv("OIDC_PROVIDER_NAME", "Pocket ID"),
		OIDCIssuerURL:    getEnv("OIDC_ISSUER_URL", ""),
		OIDCClientID:     getEnv("OIDC_CLIENT_ID", ""),
		OIDCClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:  getEnv("OIDC_REDIRECT_URL", ""),
		OIDCScopes:       getEnvAsSlice("OIDC_SCOPES", []string{"openid", "profile", "email"}),
		OIDCAuthURL:      getEnv("OIDC_AUTH_URL", ""),
		OIDCTokenURL:     getEnv("OIDC_TOKEN_URL", ""),
		OIDCUserInfoURL:  getEnv("OIDC_USER_INFO_URL", ""),

		// CORS defaults - includes Next.js dev server (3000) and production
		CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),

		// Default admin
		DefaultAdminEmail:    getEnv("DEFAULT_ADMIN_EMAIL", "admin@traefikx.local"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "changeme"),

		// Traefik HTTP Provider
		TraefikProviderToken: getEnv("TRAEFIK_PROVIDER_TOKEN", "change-me-in-production-traefik-token"),
	}

	// Validate JWT secret length
	if len(config.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters long")
	}

	AppConfig = config
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
