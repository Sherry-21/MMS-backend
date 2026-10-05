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
	"maiora-backend-service/internal/middleware"
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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. Load Application Configuration
	cfg := config.Load()
	log.Printf("Starting %s in %s mode...\n", cfg.App.Name, cfg.App.Env)

	// 2. Initialize Database Connection (GORM + PostgreSQL / Neon DB)
	db, err := database.NewConnection(&cfg.Database, cfg.App.Env)
	if err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL database: %v\n", err)
		log.Println("Server running in offline/fallback mode (mock repository will be used).")
	}

	// 3. Initialize Repositories (Data Access Layer)
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	importBatchRepo := repository.NewImportBatchRepository(db)
	stockOpnameRepo := repository.NewStockOpnameRepository(db)

	// 4. Initialize Services (Business Logic Layer)
	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessExpirationMinutes)*time.Minute,
		time.Duration(cfg.JWT.RefreshExpirationDays)*24*time.Hour,
	)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	importBatchService := service.NewImportBatchService(importBatchRepo)
	stockOpnameService := service.NewStockOpnameService(stockOpnameRepo)

	// 5. Initialize Handlers and Middlewares (Transport Layer)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret, userRepo)
	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	importBatchHandler := handler.NewImportBatchHandler(importBatchService)
	stockOpnameHandler := handler.NewStockOpnameHandler(stockOpnameService)

	// 6. Initialize Chi Router
	router := handler.NewRouter(&handler.RouterConfig{
		AuthHandler:        authHandler,
		AuthMiddleware:     authMiddleware,
		CategoryHandler:    categoryHandler,
		ProductHandler:     productHandler,
		ImportBatchHandler: importBatchHandler,
		StockOpnameHandler: stockOpnameHandler,
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
