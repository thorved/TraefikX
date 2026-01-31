package static

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	staticPath := getStaticPath()
	if staticPath != "" {
		log.Printf("Serving static files from: %s", staticPath)
		// Check if static directory exists
		if _, err := os.Stat(staticPath); !os.IsNotExist(err) {

			// Dynamically serve files and directories
			entries, err := os.ReadDir(staticPath)
			if err == nil {
				for _, entry := range entries {
					name := entry.Name()
					if name == "index.html" {
						continue
					}

					fullPath := filepath.Join(staticPath, name)
					if entry.IsDir() {
						r.Static("/"+name, fullPath)
					} else {
						r.StaticFile("/"+name, fullPath)
					}
				}
			}

			// Serve index.html for root path
			r.GET("/", func(c *gin.Context) {
				c.File(filepath.Join(staticPath, "index.html"))
			})

			// Return 404 page for all unmatched routes
			r.NoRoute(func(c *gin.Context) {
				// Try 404.html first
				p404 := filepath.Join(staticPath, "404.html")
				if _, err := os.Stat(p404); err == nil {
					c.File(p404)
					return
				}

				// Try 404/index.html
				p404Index := filepath.Join(staticPath, "404", "index.html")
				if _, err := os.Stat(p404Index); err == nil {
					c.File(p404Index)
					return
				}

				// Fallback to JSON
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
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
