package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"maiora-backend-service/config"
	"maiora-backend-service/internal/model"
)

// NewConnection establishes a GORM connection to PostgreSQL (Neon DB compatible)
func NewConnection(cfg *config.DatabaseConfig, appEnv string) (*gorm.DB, error) {
	var dsn string
	if cfg.URL != "" {
		dsn = cfg.URL
	} else {
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "require"
		}
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			cfg.Host,
			cfg.User,
			cfg.Password,
			cfg.Name,
			cfg.Port,
			sslMode,
		)
	}

	// Configure GORM logger level based on environment
	logLevel := logger.Warn
	if appEnv == "development" {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pooling via underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve generic database interface: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection with Ping
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL (Neon DB) database")
	return db, nil
}

// NewPostgresConnection is an alias for NewConnection
func NewPostgresConnection(cfg *config.DatabaseConfig, appEnv string) (*gorm.DB, error) {
	return NewConnection(cfg, appEnv)
}

// NewMySQLConnection is retained for backwards compatibility and forwards to NewConnection
func NewMySQLConnection(cfg *config.DatabaseConfig, appEnv string) (*gorm.DB, error) {
	return NewConnection(cfg, appEnv)
}

// AutoMigrate runs GORM schema migrations for all application entities
func AutoMigrate(db *gorm.DB) error {
	log.Println("Applying schema migrations...")

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.ProductVariant{},
		&model.ImportBatch{},
		&model.ImportBatchItem{},
		&model.InventoryBatch{},
		&model.StockOpname{},
		&model.StockOpnameItem{},
		&model.RefreshToken{},
	); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Println("Schema migrations applied successfully.")
	return nil
}
