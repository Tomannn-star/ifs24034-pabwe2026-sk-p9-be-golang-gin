package main

import (
	"fmt"
	"log"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.AppKey)
	authHandler := auth.NewHandler(authSvc)

	postRepo := post.NewRepository(db)
	postSvc := post.NewService(postRepo)
	postHandler := post.NewHandler(postSvc)

	openAIClient := ai.NewOpenAIClient(cfg.OpenAIEndpoint, cfg.OpenAIKey, cfg.OpenAIModel)
	searchClient := ai.NewSearchClient(cfg.SearchEngineEndpoint, "")
	aiSvc := ai.NewService(openAIClient, searchClient)
	aiHandler := ai.NewHandler(aiSvc)

	r := server.SetupRouter(cfg, authHandler, postHandler, aiHandler)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Delcom Blog API server running on port %s", cfg.AppPort)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
