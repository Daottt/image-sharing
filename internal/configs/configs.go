package configs

import (
	"os"
)

type Config struct {
	DatabaseURL     string
	Address         string
	SSOAddress      string
	ImagesDirectory string
	SchemaPath      string
}

func NewConfig() (config Config) {
	config.DatabaseURL = os.Getenv("DATABASE_URL")
	if config.DatabaseURL == "" {
		panic("DATABASE_URL not specified in .env") // postgres://user:password@host:port/database?sslmode=disable
	}
	config.Address = os.Getenv("ADDRESS")
	if config.Address == "" {
		config.Address = "localhost:8080"
	}
	config.SSOAddress = os.Getenv("SSO_ADDRESS")
	if config.SSOAddress == "" {
		config.SSOAddress = "localhost:8081"
	}
	config.ImagesDirectory = os.Getenv("IMAGES_DIRECTORY")
	if config.ImagesDirectory == "" {
		config.ImagesDirectory = "images"
	}
	config.SchemaPath = os.Getenv("SCHEMA_PATH")
	if config.SchemaPath == "" {
		config.SchemaPath = "../internal/db/schema.sql"
	}
	return config
}
