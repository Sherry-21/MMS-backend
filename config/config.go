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
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
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
			Host:            utils.GetEnv("DB_HOST", "localhost"),
			Port:            utils.GetEnv("DB_PORT", "3306"),
			User:            utils.GetEnv("DB_USER", "root"),
			Password:        utils.GetEnv("DB_PASSWORD", "password"),
			Name:            utils.GetEnv("DB_NAME", "maiora"),
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
	}
}
