package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

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
	initLog []string
	once    sync.Once
)

func initialize() {
	defer func() {
		if r := recover(); r != nil {
			initErr = fmt.Errorf("panic in initialization: %v", r)
			initLog = append(initLog, fmt.Sprintf("panic: %v", r))
		}
	}()

	start := time.Now()
	initLog = append(initLog, fmt.Sprintf("start init at %v", start))

	gin.SetMode(gin.ReleaseMode)
	cfg := config.LoadConfig()

	// In serverless, always force sqlite in-memory with shared cache
	cfg.DBConnection = "sqlite"
	cfg.DBDatabase = "file:memdb1?mode=memory&cache=shared"
	initLog = append(initLog, "configured sqlite in-memory shared cache")

	t0 := time.Now()
	db, err := database.Connect(cfg)
	if err != nil {
		initErr = fmt.Errorf("database.Connect failed: %w", err)
		initLog = append(initLog, fmt.Sprintf("db connect failed: %v", err))
		return
	}
	initLog = append(initLog, fmt.Sprintf("db connected in %v", time.Since(t0)))

	t1 := time.Now()
	if err := database.AutoMigrate(db); err != nil {
		initErr = fmt.Errorf("database.AutoMigrate failed: %w", err)
		initLog = append(initLog, fmt.Sprintf("db migrate failed: %v", err))
		return
	}
	initLog = append(initLog, fmt.Sprintf("db migrated in %v", time.Since(t1)))

	t2 := time.Now()
	if err := database.Seed(db, cfg); err != nil {
		initErr = fmt.Errorf("database.Seed failed: %w", err)
		initLog = append(initLog, fmt.Sprintf("db seed failed: %v", err))
		return
	}
	initLog = append(initLog, fmt.Sprintf("db seeded in %v", time.Since(t2)))

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
	initLog = append(initLog, fmt.Sprintf("router setup complete in %v total", time.Since(start)))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			res, _ := json.Marshal(map[string]interface{}{
				"error": map[string]string{
					"code":    "500",
					"message": fmt.Sprintf("Panic in handler: %v", rec),
				},
				"log": initLog,
			})
			_, _ = w.Write(res)
		}
	}()

	if r.URL.Path == "/debug" || r.URL.Path == "/api/debug" {
		once.Do(initialize)
		w.Header().Set("Content-Type", "application/json")
		status := "ok"
		if initErr != nil {
			status = "error"
		}
		res, _ := json.Marshal(map[string]interface{}{
			"status": status,
			"error":  fmt.Sprintf("%v", initErr),
			"log":    initLog,
			"path":   r.URL.Path,
		})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(res)
		return
	}

	once.Do(initialize)

	if initErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(map[string]interface{}{
			"error": map[string]string{
				"code":    "500",
				"message": fmt.Sprintf("Init error: %v", initErr),
			},
			"log": initLog,
		})
		_, _ = w.Write(res)
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
