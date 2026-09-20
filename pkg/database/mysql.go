package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"maiora-backend-service/config"
	"maiora-backend-service/internal/model"
)

// NewMySQLConnection establishes a GORM connection to the MySQL database
func NewMySQLConnection(cfg *config.DatabaseConfig, appEnv string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	// Configure GORM logger level based on environment
	logLevel := logger.Warn
	if appEnv == "development" {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
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
		return nil, fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	log.Printf("Successfully connected to MySQL database: %s@%s:%s/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.Name)
	return db, nil
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
	); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Println("Schema migrations applied successfully.")
	return nil
}
