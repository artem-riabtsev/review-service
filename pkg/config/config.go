package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type Config struct {
	Port int
	DB   DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	portFlag := pflag.Int("port", 0, "HTTP server port")
	pflag.Parse()

	// App port
	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	if *portFlag != 0 {
		port = *portFlag
	}

	// DB port
	dbPort := 5432
	if envDBPort := os.Getenv("POSTGRES_PORT"); envDBPort != "" {
		if p, err := strconv.Atoi(envDBPort); err == nil {
			dbPort = p
		}
	}

	return &Config{
		Port: port,
		DB: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("POSTGRES_USER", "review_user"),
			Password: getEnv("POSTGRES_PASSWORD", "review_password"),
			DBName:   getEnv("POSTGRES_DB", "review_service"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}