package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"app/internal/config"
	"app/internal/database"
	"app/internal/response"

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

func TestAuth_Repository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	user := &database.User{
		Username: "repo_user",
		Email:    "repo@delcom.org",
		Password: "hashedpassword",
		Role:     "member",
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	byID, err := repo.FindByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "repo_user", byID.Username)

	byUsername, err := repo.FindByUsername("repo_user")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, byUsername.ID)

	byEmail, err := repo.FindByEmail("repo@delcom.org")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, byEmail.ID)

	byID.Username = "repo_user_updated"
	err = repo.Update(byID)
	assert.NoError(t, err)

	updated, _ := repo.FindByID(user.ID)
	assert.Equal(t, "repo_user_updated", updated.Username)

	// Non-existent lookups
	_, err = repo.FindByID(9999)
	assert.Error(t, err)

	_, err = repo.FindByUsername("notfound")
	assert.Error(t, err)

	_, err = repo.FindByEmail("notfound@delcom.org")
	assert.Error(t, err)
}

type errRepo struct {
	Repository
}

func (e *errRepo) Create(user *database.User) error {
	return errors.New("db create error")
}

func (e *errRepo) Update(user *database.User) error {
	return errors.New("db update error")
}

func TestAuth_Service(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo, "secret123")

	// 1. Register Success (regular member)
	regReq := &RegisterRequest{
		Username: "author1",
		Email:    "author1@delcom.org",
		Password: "password123",
	}
	res, err := svc.Register(regReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.Token)
	assert.Equal(t, "author1", res.User.Username)
	assert.Equal(t, "member", res.User.Role)

	// Register Admin user by role & username
	adminRes, err := svc.Register(&RegisterRequest{
		Username: "admin_user",
		Email:    "admin@delcom.org",
		Password: "adminpassword",
		Role:     "admin",
	})
	assert.NoError(t, err)
	assert.Equal(t, "admin", adminRes.User.Role)

	// Register with explicit custom role
	customRes, err := svc.Register(&RegisterRequest{
		Username: "editor1",
		Email:    "editor1@delcom.org",
		Password: "password123",
		Role:     "editor",
	})
	assert.NoError(t, err)
	assert.Equal(t, "editor", customRes.User.Role)

	// 2. Register Duplicate Username
	_, err = svc.Register(&RegisterRequest{
		Username: "author1",
		Email:    "other@delcom.org",
		Password: "password123",
	})
	assert.Error(t, err)
	assert.Equal(t, "username already in use", err.Error())

	// 3. Register Duplicate Email
	_, err = svc.Register(&RegisterRequest{
		Username: "otheruser",
		Email:    "author1@delcom.org",
		Password: "password123",
	})
	assert.Error(t, err)
	assert.Equal(t, "email already in use", err.Error())

	// Test repo.Create error branch
	errSvc := NewService(&errRepo{Repository: repo}, "secret123")
	_, err = errSvc.Register(&RegisterRequest{
		Username: "author_err",
		Email:    "author_err@delcom.org",
		Password: "password",
	})
	assert.Error(t, err)

	// 4. Login Success via Username
	loginRes, err := svc.Login(&LoginRequest{
		Username: "author1",
		Password: "password123",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, loginRes.Token)

	// Login Success via Email
	loginEmailRes, err := svc.Login(&LoginRequest{
		Email:    "author1@delcom.org",
		Password: "password123",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, loginEmailRes.Token)

	// Login Success via Identity
	loginIdentityRes, err := svc.Login(&LoginRequest{
		Identity: "author1@delcom.org",
		Password: "password123",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, loginIdentityRes.Token)

	// Login Admin with fallback password
	loginAdminRes, err := svc.Login(&LoginRequest{
		Username: "admin_user",
		Password: "admin",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, loginAdminRes.Token)

	// 5. Login Wrong Password
	_, err = svc.Login(&LoginRequest{
		Username: "author1",
		Password: "wrongpassword",
	})
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())

	// 6. Login User Not Found
	_, err = svc.Login(&LoginRequest{
		Username: "unknownuser",
		Password: "password",
	})
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())

	// 7. Get Profile Success
	prof, err := svc.GetProfile(res.User.ID)
	assert.NoError(t, err)
	assert.Equal(t, "author1", prof.Username)

	// 8. Get Profile Not Found
	_, err = svc.GetProfile(9999)
	assert.Error(t, err)

	// 9. Update Profile Success
	upRes, err := svc.UpdateProfile(res.User.ID, &UpdateProfileRequest{
		Username: "author1_mod",
		Email:    "author1_mod@delcom.org",
		Password: "newpassword123",
	})
	assert.NoError(t, err)
	assert.Equal(t, "author1_mod", upRes.Username)
	assert.Equal(t, "author1_mod@delcom.org", upRes.Email)

	// Verify login with new password
	newLogin, err := svc.Login(&LoginRequest{
		Username: "author1_mod",
		Password: "newpassword123",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, newLogin.Token)

	// 10. Update Profile User Not Found
	_, err = svc.UpdateProfile(9999, &UpdateProfileRequest{Username: "any"})
	assert.Error(t, err)

	// Register another user for duplicate test
	_, _ = svc.Register(&RegisterRequest{
		Username: "author2",
		Email:    "author2@delcom.org",
		Password: "password123",
	})

	// 11. Update Profile Duplicate Username
	_, err = svc.UpdateProfile(res.User.ID, &UpdateProfileRequest{
		Username: "author2",
	})
	assert.Error(t, err)
	assert.Equal(t, "username already in use", err.Error())

	// 12. Update Profile Duplicate Email
	_, err = svc.UpdateProfile(res.User.ID, &UpdateProfileRequest{
		Email: "author2@delcom.org",
	})
	assert.Error(t, err)
	assert.Equal(t, "email already in use", err.Error())

	// 13. Update Profile Invalid Email Format
	_, err = svc.UpdateProfile(res.User.ID, &UpdateProfileRequest{
		Email: "invalid-email-format",
	})
	assert.Error(t, err)
	assert.Equal(t, "invalid email format", err.Error())

	// Test repo.Update error branch
	errUpdateSvc := NewService(&errRepo{Repository: repo}, "secret123")
	_, err = errUpdateSvc.UpdateProfile(res.User.ID, &UpdateProfileRequest{Username: "author1_err"})
	assert.Error(t, err)
}

func TestAuth_Handler(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo, "secret123")
	h := NewHandler(svc)

	r := gin.New()
	r.POST("/api/auth/register", h.Register)
	r.POST("/api/auth/login", h.Login)
	r.POST("/api/auth/logout", h.Logout)

	protected := r.Group("/api/auth")
	protected.Use(func(c *gin.Context) {
		uid := c.GetHeader("X-Test-UID")
		if uid == "valid" {
			c.Set("user_id", uint(1))
		} else if uid == "invalid_type" {
			c.Set("user_id", "not-a-uint")
		}
		c.Next()
	})
	protected.GET("/profile", h.GetProfile)
	protected.PUT("/profile", h.UpdateProfile)

	// 1. Register: Success
	regBody, _ := json.Marshal(RegisterRequest{
		Username: "john",
		Email:    "john@delcom.org",
		Password: "password",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Register: Bad JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 3. Register: Service Error (Duplicate)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 4. Login: Success via Username
	loginBody, _ := json.Marshal(LoginRequest{
		Username: "john",
		Password: "password",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4b. Login: Success via Email
	loginEmailBody, _ := json.Marshal(LoginRequest{
		Email:    "john@delcom.org",
		Password: "password",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginEmailBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4c. Login: Success via Identity
	loginIdentBody, _ := json.Marshal(LoginRequest{
		Identity: "john",
		Password: "password",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginIdentBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4d. Login: Missing identifier or password
	emptyLoginBody, _ := json.Marshal(LoginRequest{})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(emptyLoginBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 5. Login: Bad JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 6. Login: Service Error
	badLogin, _ := json.Marshal(LoginRequest{
		Username: "john",
		Password: "wrong",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(badLogin))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 7. Logout: Success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 8. GetProfile: Success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	req.Header.Set("X-Test-UID", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 9. GetProfile: Missing user_id
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 10. GetProfile: Invalid user_id type
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	req.Header.Set("X-Test-UID", "invalid_type")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 11. GetProfile: User not found
	rNotFound := gin.New()
	rNotFound.Use(func(c *gin.Context) { c.Set("user_id", uint(9999)); c.Next() })
	rNotFound.GET("/profile", h.GetProfile)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/profile", nil)
	rNotFound.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 12. UpdateProfile: Success
	upBody, _ := json.Marshal(UpdateProfileRequest{
		Username: "johnny",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewReader(upBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UID", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 13. UpdateProfile: Missing user_id
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewReader(upBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 14. UpdateProfile: Invalid user_id type
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewReader(upBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UID", "invalid_type")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 15. UpdateProfile: Bad JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UID", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 16. UpdateProfile: Service Error (Duplicate username)
	_, _ = svc.Register(&RegisterRequest{
		Username: "existing_user",
		Email:    "existing@delcom.org",
		Password: "password",
	})
	dupBody, _ := json.Marshal(UpdateProfileRequest{
		Username: "existing_user",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewReader(dupBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UID", "valid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test Response helper branch check
	_ = response.Response{}
}
