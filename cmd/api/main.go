package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"maiora-backend-service/config"
	"maiora-backend-service/internal/handler"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/internal/service"
	"maiora-backend-service/pkg/database"
)

// @title Maiora System Backend API
// @version 1.0
// @description Robust backend API service for Maiora System built with Go and Chi router.
// @termsOfService http://swagger.io/terms/

// @contact.name Maiora Support Team
// @contact.email support@maiora.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

func main() {
	// 1. Load Application Configuration
	cfg := config.Load()
	log.Printf("Starting %s in %s mode...\n", cfg.App.Name, cfg.App.Env)

	// 2. Initialize Database Connection (GORM + MySQL)
	db, err := database.NewMySQLConnection(&cfg.Database, cfg.App.Env)
	if err != nil {
		log.Printf("Warning: Failed to connect to MySQL database: %v\n", err)
		log.Println("Server running in offline/fallback mode (mock repository will be used).")
	}

	// 3. Initialize Repositories (Data Access Layer)
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	importBatchRepo := repository.NewImportBatchRepository(db)

	// 4. Initialize Services (Business Logic Layer)
	authService := service.NewAuthService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	importBatchService := service.NewImportBatchService(importBatchRepo)

	// 5. Initialize Handlers (Transport Layer)
	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	importBatchHandler := handler.NewImportBatchHandler(importBatchService)

	// 6. Initialize Chi Router
	router := handler.NewRouter(&handler.RouterConfig{
		AuthHandler:        authHandler,
		CategoryHandler:    categoryHandler,
		ProductHandler:     productHandler,
		ImportBatchHandler: importBatchHandler,
		AppConfig:          cfg,
	})

	// 7. Configure HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 8. Start HTTP Server in background goroutine
	go func() {
		log.Printf("Server listening on http://localhost:%s\n", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// 9. Graceful Shutdown listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited gracefully.")
}
