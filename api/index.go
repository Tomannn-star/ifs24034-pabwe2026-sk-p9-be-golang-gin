package handler

import (
	"net/http"
	"os"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/server"

	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	cfg := config.LoadConfig()

	// On Vercel serverless environment, use sqlite in /tmp for persistent session
	if os.Getenv("VERCEL") != "" || cfg.DBConnection != "pgsql" {
		cfg.DBConnection = "sqlite"
		cfg.DBDatabase = "/tmp/blog.db"
	}

	db, err := database.Connect(cfg)
	if err != nil {
		cfg.DBConnection = "sqlite"
		cfg.DBDatabase = ":memory:"
		db, _ = database.Connect(cfg)
	}

	_ = database.AutoMigrate(db)
	_ = database.Seed(db, cfg)

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

	app = server.SetupRouter(cfg, authHandler, postHandler, aiHandler)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
