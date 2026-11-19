package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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

	// In serverless environment (Vercel), use SQLite unless an external DB host is explicitly provided
	if cfg.DBHost == "127.0.0.1" || cfg.DBHost == "localhost" || cfg.DBHost == "" || os.Getenv("DB_HOST") == "" {
		cfg.DBConnection = "sqlite"
		if runtime.GOOS == "windows" {
			cfg.DBDatabase = filepath.Join(os.TempDir(), "delcom_blog.db")
		} else {
			cfg.DBDatabase = "/tmp/delcom_blog.db"
		}
	}

	db, err := database.Connect(cfg)
	if err != nil {
		cfg.DBConnection = "sqlite"
		if runtime.GOOS == "windows" {
			cfg.DBDatabase = filepath.Join(os.TempDir(), "delcom_blog.db")
		} else {
			cfg.DBDatabase = "/tmp/delcom_blog.db"
		}
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
	defer func() {
		if rec := recover(); rec != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"error":{"code":"500","message":"Panic in handler: %v"}}`, rec)))
		}
	}()

	once.Do(initialize)

	if initErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":{"code":"500","message":"Initialization error: %s"}}`, initErr.Error())))
		return
	}

	if app == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":"500","message":"Router is not initialized"}}`))
		return
	}

	app.ServeHTTP(w, r)
}
