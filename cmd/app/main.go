package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"image-sharing/internal/configs"
	"image-sharing/internal/routes"
)

const envPath = "../../.env"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	err := godotenv.Load(envPath)
	if err != nil {
		slog.Debug(".env file not found")
	}

	config := configs.NewConfig()

	if err = os.MkdirAll(config.ImagesDirectory, 0755); err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		panic(err)
	}

	SchemaPath := config.SchemaPath
	schema, err := os.ReadFile(SchemaPath)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(string(schema))
	if err != nil {
		panic(err)
	}

	router := routes.SetupRouter(db, config)

	slog.Info(fmt.Sprintf("Server started on: http://%s", config.Address))
	err = http.ListenAndServe(config.Address, router)
	if err != nil {
		slog.Error("Failed to start server", slog.Any("err", err))
	}
}
