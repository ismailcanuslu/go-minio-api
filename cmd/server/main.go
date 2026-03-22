package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"go-file-microservice/internal/api"
	"go-file-microservice/internal/config"
	"go-file-microservice/internal/storage"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	minioService, err := storage.NewMinIOService(
		ctx,
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOUseSSL,
		cfg.MinIOBucket,
	)
	if err != nil {
		log.Fatalf("storage bootstrap failed: %v", err)
	}

	controller := api.NewController(minioService)
	router := api.NewRouter(controller)
	server := api.NewHTTPServer(cfg.ServerPort, router)

	log.Printf("server listening on :%s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
