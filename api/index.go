package handler

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/server"

	"github.com/gin-gonic/gin"
)

var (
	app     *gin.Engine
	initErr error
	once    sync.Once
)

func initialize() {
	defer func() {
		if r := recover(); r != nil {
			initErr = fmt.Errorf("panic in initialization: %v", r)
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	cfg := config.LoadConfig()

	// On Vercel serverless environment, use in-memory SQLite
	if os.Getenv("VERCEL") != "" || cfg.DBConnection != "pgsql" {
		cfg.DBConnection = "sqlite"
		cfg.DBDatabase = ":memory:"
	}

	db, err := database.Connect(cfg)
	if err != nil {
		cfg.DBConnection = "sqlite"
		cfg.DBDatabase = ":memory:"
		db, err = database.Connect(cfg)
		if err != nil {
			initErr = fmt.Errorf("failed to connect database: %w", err)
			return
		}
	}

	if err := database.AutoMigrate(db); err != nil {
		initErr = fmt.Errorf("failed to migrate database: %w", err)
		return
	}

	if err := database.Seed(db, cfg); err != nil {
		initErr = fmt.Errorf("failed to seed database: %w", err)
		return
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

	app = server.SetupRouter(cfg, authHandler, postHandler, aiHandler)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initialize)

	if initErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"status":"error","message":"%s"}`, initErr.Error())))
		return
	}

	app.ServeHTTP(w, r)
}
