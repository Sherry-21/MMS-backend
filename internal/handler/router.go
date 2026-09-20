package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"maiora-backend-service/config"
	_ "maiora-backend-service/docs"
	"maiora-backend-service/pkg/utils"
)

// RouterConfig holds handler dependencies for router setup
type RouterConfig struct {
	AuthHandler        *AuthHandler
	ImportBatchHandler *ImportBatchHandler
	CategoryHandler    *CategoryHandler
	ProductHandler     *ProductHandler
	AppConfig          *config.Config
}

// NewRouter initializes and returns a configured Chi HTTP router
func NewRouter(cfg *RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Base Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS Setup
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AppConfig.CORS.AllowedOrigins,
		AllowedMethods:   cfg.AppConfig.CORS.AllowedMethods,
		AllowedHeaders:   cfg.AppConfig.CORS.AllowedHeaders,
		AllowCredentials: cfg.AppConfig.CORS.AllowCredentials,
		MaxAge:           300, // Maximum value not ignored by any major browser
	}))

	// Swagger UI Route
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Root & Health Check Endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w, http.StatusOK, cfg.AppConfig.App.Name+" is running", map[string]string{
			"environment": cfg.AppConfig.App.Env,
			"version":     "1.0.0",
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w, http.StatusOK, "Service is healthy", map[string]string{
			"status": "UP",
		})
	})

	// API Routes Group
	r.Route("/api/v1", func(api chi.Router) {
		// Authentication Routes
		if cfg.AuthHandler != nil {
			api.Route("/auth", func(auth chi.Router) {
				auth.Post("/register", cfg.AuthHandler.Register)
				auth.Post("/login", cfg.AuthHandler.Login)
			})
		}

		// Category Routes
		if cfg.CategoryHandler != nil {
			api.Route("/categories", func(cat chi.Router) {
				cat.Post("/", cfg.CategoryHandler.CreateCategory)
				cat.Get("/", cfg.CategoryHandler.ListCategories)
			})
		}

		// Product & Variant Routes
		if cfg.ProductHandler != nil {
			api.Route("/products", func(prod chi.Router) {
				prod.Post("/", cfg.ProductHandler.CreateProduct)
				prod.Get("/", cfg.ProductHandler.ListProducts)
				prod.Get("/{id}", cfg.ProductHandler.GetProductByID)
				prod.Post("/{id}/variants", cfg.ProductHandler.CreateVariant)
			})

			// Product Variants Global List (for Import Batch dropdowns)
			api.Get("/product-variants", cfg.ProductHandler.ListVariants)
		}

		// Import Batch & Inventory Management Routes
		if cfg.ImportBatchHandler != nil {
			api.Route("/import-batches", func(ib chi.Router) {
				ib.Post("/", cfg.ImportBatchHandler.CreateImportBatch)
			})
		}
	})

	return r
}
