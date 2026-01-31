package static

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	staticPath := getStaticPath()
	if staticPath != "" {
		log.Printf("Serving static files from: %s", staticPath)
		// Check if static directory exists
		if _, err := os.Stat(staticPath); !os.IsNotExist(err) {
			// Serve Next.js static files
			r.Static("/_next", filepath.Join(staticPath, "_next"))
			r.StaticFile("/favicon.ico", filepath.Join(staticPath, "favicon.ico"))

			// Serve page directories (login, dashboard, etc.)
			r.Static("/login", filepath.Join(staticPath, "login"))
			r.Static("/dashboard", filepath.Join(staticPath, "dashboard"))
			r.Static("/users", filepath.Join(staticPath, "users"))
			r.Static("/profile", filepath.Join(staticPath, "profile"))
			r.Static("/auth", filepath.Join(staticPath, "auth"))

			// Serve index.html for root path
			r.GET("/", func(c *gin.Context) {
				c.File(filepath.Join(staticPath, "index.html"))
			})

			// Serve index.html for all non-API, non-static routes (SPA fallback)
			r.NoRoute(func(c *gin.Context) {
				path := c.Request.URL.Path
				if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/_next/") {
					c.File(filepath.Join(staticPath, "index.html"))
				} else {
					c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				}
			})
		} else {
			log.Printf("Warning: Static path not found: %s", staticPath)
		}
	} else {
		log.Println("Warning: No static files path configured")
	}
}

func getStaticPath() string {
	// Try multiple possible locations for the Next.js frontend build
	// Next.js with output: 'export' creates an 'out' directory
	possiblePaths := []string{
		// Same directory as binary (production)
		func() string {
			execPath, err := os.Executable()
			if err != nil {
				return ""
			}
			return filepath.Join(filepath.Dir(execPath), "frontend", "out")
		}(),
		// Parent directory of backend (development from backend/)
		"../frontend/out",
		// Current working directory
		"./frontend/out",
		// Absolute path from project root
		"./../frontend/out",
	}

	for _, path := range possiblePaths {
		if path == "" {
			continue
		}
		// Check if index.html exists in this path
		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return path
		}
	}

	return ""
}
