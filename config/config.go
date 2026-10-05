package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"

	"maiora-backend-service/pkg/utils"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	CORS     CORSConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL             string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type JWTConfig struct {
	Secret                  string
	AccessExpirationMinutes int
	RefreshExpirationDays   int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error reading it, using system environment variables")
	}

	return &Config{
		App: AppConfig{
			Name: utils.GetEnv("APP_NAME", "Maiora System Backend"),
			Env:  utils.GetEnv("APP_ENV", "development"),
		},
		Server: ServerConfig{
			Port:         utils.GetEnv("PORT", "8080"),
			ReadTimeout:  utils.GetEnvAsDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout: utils.GetEnvAsDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  utils.GetEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:             utils.GetEnv("DATABASE_URL", utils.GetEnv("DB_URL", "")),
			Host:            utils.GetEnv("DB_HOST", "localhost"),
			Port:            utils.GetEnv("DB_PORT", "5432"),
			User:            utils.GetEnv("DB_USER", "postgres"),
			Password:        utils.GetEnv("DB_PASSWORD", "password"),
			Name:            utils.GetEnv("DB_NAME", "neondb"),
			SSLMode:         utils.GetEnv("DB_SSLMODE", "require"),
			MaxOpenConns:    utils.GetEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    utils.GetEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: utils.GetEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		CORS: CORSConfig{
			AllowedOrigins:   utils.GetEnvAsSlice("ALLOWED_ORIGINS", []string{"*"}, ","),
			AllowedMethods:   utils.GetEnvAsSlice("ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, ","),
			AllowedHeaders:   utils.GetEnvAsSlice("ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}, ","),
			AllowCredentials: utils.GetEnvAsBool("ALLOW_CREDENTIALS", true),
		},
		JWT: JWTConfig{
			Secret:                  utils.GetEnv("JWT_SECRET", utils.GetEnv("SECRET_KEY", "maiora-super-secret-jwt-key-change-in-production")),
			AccessExpirationMinutes: utils.GetEnvAsInt("JWT_ACCESS_EXPIRATION_MINUTES", 15),
			RefreshExpirationDays:   utils.GetEnvAsInt("JWT_REFRESH_EXPIRATION_DAYS", 7),
		},
	}
}
