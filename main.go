package main

import (
	"fmt"
	"log"
	"os"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	cfg := config.LoadConfig()

	// Connect to database (try PostgreSQL, fallback to SQLite if unavailable)
	db, err := database.Connect(cfg)
	if err != nil {
		log.Printf("Connecting to %s database failed: %v. Falling back to SQLite...", cfg.DBConnection, err)
		cfg.DBConnection = "sqlite"
		if os.PathSeparator == '/' {
			cfg.DBDatabase = "/tmp/p9_blog.db"
		} else {
			cfg.DBDatabase = "p9_blog.db"
		}
		db, err = database.Connect(cfg)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
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

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.AppPort
	}
	if port == "" {
		port = "8081"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Delcom Blog API server running on port %s", port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
