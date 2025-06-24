package configs

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	DatabaseURL     string
	Address         string
	SecretKey       string
	ImagesDirectory string
	SchemaPath      string
}

const minSecretKeySize = 32

func NewConfig() (config Config) {
	config.DatabaseURL = os.Getenv("DATABASE_URL")
	if config.DatabaseURL == "" {
		panic("DATABASE_URL not specified in .env") // postgres://user:password@host:port/database?sslmode=disable
	}
	config.Address = os.Getenv("ADDRESS")
	if config.Address == "" {
		config.Address = "localhost:8080"
	}
	config.SecretKey = os.Getenv("SECRET_KEY")
	if config.SecretKey == "" || len(config.SecretKey) < minSecretKeySize {
		config.SecretKey = "01234567890123456789012345678901"
		slog.Error(fmt.Sprintf("SECRET_KEY must be at least %d characters. Using default secret key", minSecretKeySize))
	}
	config.ImagesDirectory = os.Getenv("IMAGES_DIRECTORY")
	if config.ImagesDirectory == "" {
		config.ImagesDirectory = "images"
	}
	config.SchemaPath = os.Getenv("SCHEMA_PATH")
	if config.SchemaPath == "" {
		config.SchemaPath = "../../internal/db/schema.sql"
	}
	return config
}
