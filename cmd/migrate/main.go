package main

import (
	"log"
	"os"

	"maiora-backend-service/config"
	"maiora-backend-service/pkg/database"
)

func main() {
	log.Println("==================================================")
	log.Println("        Maiora Database Migration Tool            ")
	log.Println("==================================================")

	// 1. Load Configuration
	cfg := config.Load()

	// 2. Connect to MySQL
	db, err := database.NewMySQLConnection(&cfg.Database, cfg.App.Env)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Execute Schema Migration
	log.Println("Executing database schema migrations...")
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("❌ Migration failed: %v\n", err)
		os.Exit(1)
	}

	log.Println("✅ Database migrations completed successfully!")
	os.Exit(0)
}
