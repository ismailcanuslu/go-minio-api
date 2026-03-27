package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	)
	if err != nil {
		log.Fatalf("storage bootstrap failed: %v", err)
	}

	controller := api.NewController(minioService)
	router := api.NewRouter(controller)
	server := api.NewHTTPServer(cfg.ServerPort, router)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited cleanly")
}
