package post

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestDB(t *testing.T) *gorm.DB {
	cfg := &config.Config{
		DBConnection: "sqlite",
		DBDatabase:   ":memory:",
	}
	db, err := database.Connect(cfg)
	assert.NoError(t, err)
	err = database.AutoMigrate(db)
	assert.NoError(t, err)
	return db
}

func TestPost_ReadingTimeAndSlug(t *testing.T) {
	// Reading time
	assert.Equal(t, 1, CalculateReadingTime(""))
	assert.Equal(t, 1, CalculateReadingTime("hello world"))
	assert.Equal(t, 2, CalculateReadingTime(stringsRepeat("word ", 250)))

	// Slug
	assert.Equal(t, "post", GenerateSlug(""))
	assert.Equal(t, "post", GenerateSlug("!!! ???"))
	assert.Equal(t, "judul-artikel-keren", GenerateSlug("Judul Artikel Keren!"))
}

func stringsRepeat(s string, count int) string {
	res := ""
	for i := 0; i < count; i++ {
		res += s
	}
	return res
}

func TestPost_Repository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	user := &database.User{
		Username: "author_test",
		Email:    "author@delcom.org",
		Password: "password",
		Role:     "member",
	}
	_ = db.Create(user)

	now := time.Now()
	post1 := &database.Post{
		UserID:      user.ID,
		Title:       "Post Pertama",
		Slug:        "post-pertama",
		Category:    "Teknologi",
		Excerpt:     "Ringkasan",
		Content:     "Konten artikel lengkap...",
		IsPublished: true,
		PublishedAt: &now,
	}
	err := repo.Create(post1)
	assert.NoError(t, err)

	post2 := &database.Post{
		UserID:      user.ID,
		Title:       "Post Draf",
		Slug:        "post-draf",
		Category:    "Pendidikan",
		Excerpt:     "Draf ringkasan",
		Content:     "Draf konten",
		IsPublished: false,
	}
	_ = repo.Create(post2)

	// FindAllPublic
	pubPosts, err := repo.FindAllPublic("", "")
	assert.NoError(t, err)
	assert.Len(t, pubPosts, 1)

	// FindAllPublic with category filter
	catPosts, err := repo.FindAllPublic("", "Teknologi")
	assert.NoError(t, err)
	assert.Len(t, catPosts, 1)

	allCatPosts, err := repo.FindAllPublic("", "Semua")
	assert.NoError(t, err)
	assert.Len(t, allCatPosts, 1)

	// FindAllPublic with query search
	qPosts, err := repo.FindAllPublic("Pertama", "")
	assert.NoError(t, err)
	assert.Len(t, qPosts, 1)

	// FindBySlugPublic
	pPub, err := repo.FindBySlugPublic("post-pertama")
	assert.NoError(t, err)
	assert.Equal(t, "Post Pertama", pPub.Title)

	_, err = repo.FindBySlugPublic("post-draf")
	assert.Error(t, err)

	// FindAllByUserID
	userPosts, err := repo.FindAllByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, userPosts, 2)

	// FindByIDAndUserID
	pByIDUser, err := repo.FindByIDAndUserID(post1.ID, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, post1.ID, pByIDUser.ID)

	_, err = repo.FindByIDAndUserID(9999, user.ID)
	assert.Error(t, err)

	// FindByID
	pByID, err := repo.FindByID(post1.ID)
	assert.NoError(t, err)
	assert.Equal(t, post1.ID, pByID.ID)

	// FindBySlug
	pBySlug, err := repo.FindBySlug("post-pertama")
	assert.NoError(t, err)
	assert.Equal(t, post1.ID, pBySlug.ID)

	// Update
	post1.Title = "Post Pertama Dimodifikasi"
	err = repo.Update(post1)
	assert.NoError(t, err)

	// Stats
	stats, err := repo.GetStatsByUserID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats.Total)
	assert.Equal(t, int64(1), stats.TotalPublished)
	assert.Equal(t, int64(1), stats.TotalDraft)

	// Admin posts
	adminPosts, err := repo.FindAllAdminPosts()
	assert.NoError(t, err)
	assert.Len(t, adminPosts, 2)

	// Admin users
	adminUsers, err := repo.FindAllAdminUsers()
	assert.NoError(t, err)
	assert.Len(t, adminUsers, 1)
	assert.Equal(t, int64(2), adminUsers[0].TotalPosts)
	assert.Equal(t, int64(1), adminUsers[0].PublishedPosts)

	// Admin user posts
	adminUserPosts, err := repo.FindAllByUserIDAdmin(user.ID)
	assert.NoError(t, err)
	assert.Len(t, adminUserPosts, 2)

	// Delete
	err = repo.Delete(post2)
	assert.NoError(t, err)
}

func TestPost_Service(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	user := &database.User{
		Username: "author_svc",
		Email:    "authorsvc@delcom.org",
		Password: "password",
		Role:     "member",
	}
	_ = db.Create(user)

	// 1. Categories
	cats := svc.GetCategories()
	assert.Equal(t, StandardCategories, cats)

	// 2. Create Post
	isPub := true
	createReq := &CreatePostRequest{
		Title:          "Artikel AI Terkini",
		Category:       "Teknologi",
		Excerpt:        "Ringkasan artikel",
		Content:        "Ini adalah konten artikel tentang kecerdasan buatan dan perkembangannya di dunia.",
		CoverImagePath: "/images/cover.jpg",
		IsPublished:    &isPub,
	}
	res, err := svc.CreatePost(user.ID, createReq)
	assert.NoError(t, err)
	assert.Equal(t, "artikel-ai-terkini", res.Slug)
	assert.True(t, res.IsPublished)
	assert.NotNil(t, res.PublishedAt)
	assert.Equal(t, "author_svc", res.AuthorName)

	// Create duplicate title post to test slug uniqueness
	createReq2 := &CreatePostRequest{
		Title:    "Artikel AI Terkini",
		Category: "", // test default category branch
		Content:  "Konten artikel kedua...",
	}
	res2, err := svc.CreatePost(user.ID, createReq2)
	assert.NoError(t, err)
	assert.Equal(t, "artikel-ai-terkini-1", res2.Slug)
	assert.Equal(t, "Umum", res2.Category)
	assert.False(t, res2.IsPublished)

	// 3. Get Public Posts
	pubList, err := svc.GetPublicPosts("", "")
	assert.NoError(t, err)
	assert.Len(t, pubList, 1)

	// 4. Get Public Post By Slug
	pubPost, err := svc.GetPublicPostBySlug("artikel-ai-terkini")
	assert.NoError(t, err)
	assert.Equal(t, res.ID, pubPost.ID)

	_, err = svc.GetPublicPostBySlug("not-exist")
	assert.Error(t, err)

	// 5. Get Author Posts
	authorPosts, err := svc.GetAuthorPosts(user.ID)
	assert.NoError(t, err)
	assert.Len(t, authorPosts, 2)

	// 6. Get Author Post By ID
	postByID, err := svc.GetAuthorPostByID(res.ID, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, res.ID, postByID.ID)

	_, err = svc.GetAuthorPostByID(9999, user.ID)
	assert.Error(t, err)

	// 7. Update Author Post
	isPubFalse := false
	upReq := &UpdatePostRequest{
		Title:          "Artikel AI Terkini Versi 2",
		Category:       "Bisnis",
		Excerpt:        "Ringkasan baru",
		Content:        "Konten baru",
		CoverImagePath: "/new-cover.jpg",
		IsPublished:    &isPubFalse,
	}
	upRes, err := svc.UpdateAuthorPost(res.ID, user.ID, upReq)
	assert.NoError(t, err)
	assert.Equal(t, "artikel-ai-terkini-versi-2", upRes.Slug)
	assert.Equal(t, "Bisnis", upRes.Category)
	assert.False(t, upRes.IsPublished)
	assert.Nil(t, upRes.PublishedAt)

	// Toggle back to published
	isPubTrue := true
	upReq2 := &UpdatePostRequest{
		IsPublished: &isPubTrue,
	}
	upRes2, err := svc.UpdateAuthorPost(res.ID, user.ID, upReq2)
	assert.NoError(t, err)
	assert.True(t, upRes2.IsPublished)
	assert.NotNil(t, upRes2.PublishedAt)

	_, err = svc.UpdateAuthorPost(9999, user.ID, upReq)
	assert.Error(t, err)

	// 8. Toggle Publish
	tog1, err := svc.TogglePublish(res.ID, user.ID)
	assert.NoError(t, err)
	assert.False(t, tog1.IsPublished)

	tog2, err := svc.TogglePublish(res.ID, user.ID)
	assert.NoError(t, err)
	assert.True(t, tog2.IsPublished)

	_, err = svc.TogglePublish(9999, user.ID)
	assert.Error(t, err)

	// 9. Stats
	stats, err := svc.GetPostStats(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats.Total)

	// 10. Admin methods
	admUsers, err := svc.GetAdminUsers()
	assert.NoError(t, err)
	assert.Len(t, admUsers, 1)

	admPosts, err := svc.GetAdminPosts()
	assert.NoError(t, err)
	assert.Len(t, admPosts, 2)

	admUserPosts, err := svc.GetAdminUserPosts(user.ID)
	assert.NoError(t, err)
	assert.Len(t, admUserPosts, 2)

	// 11. Delete Author Post
	err = svc.DeleteAuthorPost(res2.ID, user.ID)
	assert.NoError(t, err)

	err = svc.DeleteAuthorPost(9999, user.ID)
	assert.Error(t, err)

	// Test repo errors in service
	mRepo := &mockErrRepo{Repository: repo, failCreate: true}
	errSvc := NewService(mRepo)
	_, err = errSvc.CreatePost(user.ID, createReq)
	assert.Error(t, err)

	mRepo.failCreate = false
	mRepo.failUpdate = true
	_, err = errSvc.UpdateAuthorPost(res.ID, user.ID, upReq)
	assert.Error(t, err)

	_, err = errSvc.TogglePublish(res.ID, user.ID)
	assert.Error(t, err)
}

type mockErrRepo struct {
	Repository
	failCreate bool
	failUpdate bool
}

func (m *mockErrRepo) Create(post *database.Post) error {
	if m.failCreate {
		return errors.New("db create error")
	}
	return m.Repository.Create(post)
}

func (m *mockErrRepo) Update(post *database.Post) error {
	if m.failUpdate {
		return errors.New("db update error")
	}
	return m.Repository.Update(post)
}

type mockErrPostService struct {
	Service
}

func (m *mockErrPostService) GetPublicPosts(q, category string) ([]PostResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) GetAuthorPosts(userID uint) ([]PostResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) CreatePost(userID uint, req *CreatePostRequest) (*PostResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) GetPostStats(userID uint) (*PostStatsResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) GetAdminUsers() ([]AdminUserStatsResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) GetAdminPosts() ([]PostResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) GetAdminUserPosts(userID uint) ([]PostResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) UpdateAuthorPost(id, userID uint, req *UpdatePostRequest) (*PostResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) TogglePublish(id, userID uint) (*PostResponse, error) {
	return nil, errors.New("err")
}
func (m *mockErrPostService) DeleteAuthorPost(id, userID uint) error {
	return errors.New("err")
}

func TestPost_Handler(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	user := &database.User{Username: "testauthor", Email: "testauthor@delcom.org", Password: "pass"}
	_ = db.Create(user)

	r := gin.New()
	r.GET("/api/public/categories", h.GetCategories)
	r.GET("/api/public/posts", h.GetPublicPosts)
	r.GET("/api/public/posts/:slug", h.GetPublicPostBySlug)

	authGroup := r.Group("/api/posts")
	authGroup.Use(func(c *gin.Context) {
		hVal := c.GetHeader("X-Test-Auth")
		if hVal == "valid" {
			c.Set("user_id", user.ID)
		}
		c.Next()
	})
	authGroup.GET("", h.GetAuthorPosts)
	authGroup.POST("", h.CreatePost)
	authGroup.GET("/stats", h.GetPostStats)
	authGroup.GET("/:id", h.GetAuthorPostByID)
	authGroup.PUT("/:id", h.UpdateAuthorPost)
	authGroup.PATCH("/:id/toggle-publish", h.TogglePublish)
	authGroup.DELETE("/:id", h.DeleteAuthorPost)

	admGroup := r.Group("/api/admin")
	admGroup.GET("/users", h.GetAdminUsers)
	admGroup.GET("/posts", h.GetAdminPosts)
	admGroup.GET("/users/:id/posts", h.GetAdminUserPosts)

	// 1. GetCategories: 200
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/public/categories", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. GetPublicPosts: 200
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/public/posts?q=test&category=Umum", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. CreatePost: 201
	body, _ := json.Marshal(CreatePostRequest{
		Title:    "Sample Post",
		Category: "Umum",
		Content:  "Sample content",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 4. CreatePost: Unauthorized
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 5. CreatePost: Bad JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/posts", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 6. GetPublicPostBySlug: 200 & 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/public/posts/not-found-slug", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Publish post for testing slug lookup
	_, _ = svc.TogglePublish(1, user.ID)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/public/posts/sample-post", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 7. GetAuthorPosts: 200 & Unauthorized
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 8. GetPostStats: 200 & Unauthorized
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts/stats", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts/stats", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 9. GetAuthorPostByID: 200, 401, 400, 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts/1", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts/abc", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts/9999", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 10. UpdateAuthorPost: 200, 401, 400 (bad id), 400 (bad json), 400 (not found)
	upBody, _ := json.Marshal(UpdatePostRequest{Title: "Updated title"})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/posts/1", bytes.NewReader(upBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/posts/1", bytes.NewReader(upBody))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/posts/abc", bytes.NewReader(upBody))
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/posts/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/posts/9999", bytes.NewReader(upBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 11. TogglePublish: 200, 401, 400 (bad id), 404 (not found)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/api/posts/1/toggle-publish", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/api/posts/1/toggle-publish", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/api/posts/abc/toggle-publish", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/api/posts/9999/toggle-publish", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 12. DeleteAuthorPost: 200, 401, 400 (bad id), 404 (not found)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/posts/1", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/posts/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/posts/abc", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/posts/9999", nil)
	req.Header.Set("X-Test-Auth", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 13. Admin handlers
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users/1/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users/abc/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 14. Test mock error 500 handlers
	errH := NewHandler(&mockErrPostService{Service: svc})
	rErr := gin.New()
	rErr.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })
	rErr.GET("/api/public/posts", errH.GetPublicPosts)
	rErr.GET("/api/posts", errH.GetAuthorPosts)
	rErr.POST("/api/posts", errH.CreatePost)
	rErr.GET("/api/posts/stats", errH.GetPostStats)
	rErr.GET("/api/admin/users", errH.GetAdminUsers)
	rErr.GET("/api/admin/posts", errH.GetAdminPosts)
	rErr.GET("/api/admin/users/:id/posts", errH.GetAdminUserPosts)
	rErr.PUT("/api/posts/:id", errH.UpdateAuthorPost)
	rErr.PATCH("/api/posts/:id/toggle-publish", errH.TogglePublish)
	rErr.DELETE("/api/posts/:id", errH.DeleteAuthorPost)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/public/posts", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/posts/stats", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/posts", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/users/1/posts", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/posts/1", bytes.NewReader(upBody))
	req.Header.Set("Content-Type", "application/json")
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/api/posts/1/toggle-publish", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/posts/1", nil)
	rErr.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 15. Forbidden tests (post belongs to user.ID, accessed by userID 999)
	pForbid, _ := svc.CreatePost(user.ID, &CreatePostRequest{
		Title:    "Post To Forbid",
		Category: "Umum",
		Content:  "Content forbid",
	})
	forbidPath := fmt.Sprintf("/api/posts/%d", pForbid.ID)

	rForbid := gin.New()
	rForbid.Use(func(c *gin.Context) { c.Set("user_id", uint(999)); c.Next() })
	rForbid.GET("/api/posts/:id", h.GetAuthorPostByID)
	rForbid.PUT("/api/posts/:id", h.UpdateAuthorPost)
	rForbid.PATCH("/api/posts/:id/toggle-publish", h.TogglePublish)
	rForbid.DELETE("/api/posts/:id", h.DeleteAuthorPost)

	wF := httptest.NewRecorder()
	reqF, _ := http.NewRequest(http.MethodGet, forbidPath, nil)
	rForbid.ServeHTTP(wF, reqF)
	assert.Equal(t, http.StatusForbidden, wF.Code)

	wF = httptest.NewRecorder()
	reqF, _ = http.NewRequest(http.MethodPut, forbidPath, bytes.NewReader(upBody))
	reqF.Header.Set("Content-Type", "application/json")
	rForbid.ServeHTTP(wF, reqF)
	assert.Equal(t, http.StatusForbidden, wF.Code)

	wF = httptest.NewRecorder()
	reqF, _ = http.NewRequest(http.MethodPatch, forbidPath+"/toggle-publish", nil)
	rForbid.ServeHTTP(wF, reqF)
	assert.Equal(t, http.StatusForbidden, wF.Code)

	wF = httptest.NewRecorder()
	reqF, _ = http.NewRequest(http.MethodDelete, forbidPath, nil)
	rForbid.ServeHTTP(wF, reqF)
	assert.Equal(t, http.StatusForbidden, wF.Code)
}
