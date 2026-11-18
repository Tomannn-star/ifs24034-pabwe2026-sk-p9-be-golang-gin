package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func generateTestToken(userID uint, username, role, secret string) string {
	claims := &middleware.JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func setupTestRouter(t *testing.T) (*gin.Engine, *config.Config, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&database.User{}, &database.Post{})
	assert.NoError(t, err)

	cfg := &config.Config{
		AppPort: "8081",
		AppKey:  "testsecretkeyforroutertesting123456",
	}

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.AppKey)
	authH := auth.NewHandler(authSvc)

	postRepo := post.NewRepository(db)
	postSvc := post.NewService(postRepo)
	postH := post.NewHandler(postSvc)

	aiSvc := ai.NewService(&mockAIClient{}, nil)
	aiH := ai.NewHandler(aiSvc)

	r := SetupRouter(cfg, authH, postH, aiH)
	return r, cfg, db
}

type mockAIClient struct{}

func (m *mockAIClient) Chat(sys, user string) (string, error) {
	return `{"title":"T","category":"C","excerpt":"E","content":"K"}`, nil
}

func TestRouter_NoRoute(t *testing.T) {
	r, _, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/unknown/path", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Endpoint not found")
}

func TestRouter_SwaggerAndAPIJSON(t *testing.T) {
	r, _, _ := setupTestRouter(t)

	// GET /api.json (plain HTTP)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api.json", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "swagger")

	// GET /api.json (TLS)
	wTLS := httptest.NewRecorder()
	reqTLS, _ := http.NewRequest(http.MethodGet, "/api.json", nil)
	reqTLS.TLS = &tls.ConnectionState{}
	r.ServeHTTP(wTLS, reqTLS)
	assert.Equal(t, http.StatusOK, wTLS.Code)
	assert.Contains(t, wTLS.Body.String(), "swagger")

	// GET /swagger/index.html
	wSwag := httptest.NewRecorder()
	reqSwag, _ := http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	reqSwag.RequestURI = "/swagger/index.html"
	r.ServeHTTP(wSwag, reqSwag)
	assert.Equal(t, http.StatusOK, wSwag.Code)
}

func TestRouter_PublicAndProtectedRoutes(t *testing.T) {
	r, cfg, db := setupTestRouter(t)

	// 1. Public categories
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/public/categories", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Public posts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/public/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Protected author route without auth
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 4. Protected admin route without auth
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 5. Protected admin route with member token (forbidden)
	memberToken := generateTestToken(99, "memberuser", "member", cfg.AppKey)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// 6. Protected admin route with admin token (success)
	adminToken := generateTestToken(1, "adminuser", "admin", cfg.AppKey)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	_ = db
}
