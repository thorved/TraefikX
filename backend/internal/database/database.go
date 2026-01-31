package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/traefikx/backend/internal/config"
	"github.com/traefikx/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) (*gorm.DB, error) {
	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DatabasePath)
	if dbDir != "" && dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, err
		}
	}

	var logLevel logger.LogLevel
	if cfg.Env == "development" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}

func Migrate() error {
	if DB == nil {
		return nil
	}

	log.Println("Running database migrations...")

	// Auto-migrate models
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Router{},
		&models.RouterHostname{},
		&models.RouterMiddleware{},
		&models.Service{},
		&models.ServiceServer{},
		&models.Middleware{},
	); err != nil {
		return err
	}

	log.Println("Database migrations completed")
	return nil
}

func CreateDefaultAdmin(cfg *config.Config) error {
	if DB == nil {
		return nil
	}

	var count int64
	if err := DB.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	// Only create default admin if no users exist
	if count == 0 {
		log.Println("Creating default admin user...")

		// Hash password
		hashedPassword, err := HashPassword(cfg.DefaultAdminPassword)
		if err != nil {
			return err
		}

		admin := models.User{
			Email:           cfg.DefaultAdminEmail,
			Password:        hashedPassword,
			Role:            models.RoleAdmin,
			IsActive:        true,
			PasswordEnabled: true,
			OIDCEnabled:     false,
		}

		if err := DB.Create(&admin).Error; err != nil {
			return err
		}

		log.Printf("Default admin user created: %s", cfg.DefaultAdminEmail)
		log.Println("IMPORTANT: Please change the default admin password immediately!")
	}

	return nil
}
