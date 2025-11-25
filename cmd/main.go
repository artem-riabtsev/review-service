package main

import (
	"context"
	"log"
	"os"

	"review-service/internal/handler"
	"review-service/internal/repository"
	"review-service/internal/service"
	"review-service/pkg/config"
	"review-service/pkg/database"
	"review-service/pkg/server"
)

func main() {
	log.SetOutput(os.Stdout)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	dbPool, err := database.New(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	repo := repository.NewPostgresRepository(dbPool)

	svc := service.NewService(repo)

	router := handler.NewHandler(svc)

	server.Start(cfg.Port, router)
}