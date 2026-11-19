package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMiddleware_CORS(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Test GET
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	// Test OPTIONS
	wOpt := httptest.NewRecorder()
	reqOpt, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	r.ServeHTTP(wOpt, reqOpt)

	assert.Equal(t, http.StatusNoContent, wOpt.Code)
}

func generateTestToken(secret string, userID uint, username, email, role string, d time.Duration) string {
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestMiddleware_Auth(t *testing.T) {
	secret := "mysecret"
	r := gin.New()
	r.Use(Auth(secret))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "welcome "+c.GetString("username"))
	})

	// 1. Missing header
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// 2. Malformed header
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Basic abc")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// 3. Invalid token string
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer invalid.token.value")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)

	// 4. Valid token
	validToken := generateTestToken(secret, 1, "testuser", "test@delcom.org", "member", time.Hour)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req4.Header.Set("Authorization", "Bearer "+validToken)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
	assert.Equal(t, "welcome testuser", w4.Body.String())
}

func TestMiddleware_RequireRole(t *testing.T) {
	// 1. No role set
	r1 := gin.New()
	r1.Use(RequireRole("admin"))
	r1.GET("/admin", func(c *gin.Context) { c.String(http.StatusOK, "admin ok") })
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	r1.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusForbidden, w1.Code)

	// 2. Non-string role
	r2 := gin.New()
	r2.Use(func(c *gin.Context) { c.Set("role", 12345); c.Next() })
	r2.Use(RequireRole("admin"))
	r2.GET("/admin", func(c *gin.Context) { c.String(http.StatusOK, "admin ok") })
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)

	// 3. Role mismatch
	r3 := gin.New()
	r3.Use(func(c *gin.Context) { c.Set("role", "member"); c.Next() })
	r3.Use(RequireRole("admin"))
	r3.GET("/admin", func(c *gin.Context) { c.String(http.StatusOK, "admin ok") })
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	r3.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusForbidden, w3.Code)

	// 4. Role matched directly
	r4 := gin.New()
	r4.Use(func(c *gin.Context) { c.Set("role", "admin"); c.Next() })
	r4.Use(RequireRole("admin", "superadmin"))
	r4.GET("/admin", func(c *gin.Context) { c.String(http.StatusOK, "admin ok") })
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	r4.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// 5. Matched via username containing admin
	r5 := gin.New()
	r5.Use(func(c *gin.Context) {
		c.Set("role", "member")
		c.Set("username", "admin_tester")
		c.Next()
	})
	r5.Use(RequireRole("admin"))
	r5.GET("/admin", func(c *gin.Context) { c.String(http.StatusOK, "admin ok") })
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	r5.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusOK, w5.Code)

	// 6. Matched via email containing admin
	r6 := gin.New()
	r6.Use(func(c *gin.Context) {
		c.Set("role", "member")
		c.Set("username", "user123")
		c.Set("email", "admin_tester@delcom.org")
		c.Next()
	})
	r6.Use(RequireRole("admin"))
	r6.GET("/admin", func(c *gin.Context) { c.String(http.StatusOK, "admin ok") })
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	r6.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusOK, w6.Code)
}

func TestMiddleware_RateLimiter(t *testing.T) {
	storage := NewRateLimiterStorage(2, 50*time.Millisecond)
	r := gin.New()
	r.Use(RateLimiterWithStorage(storage))
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	// Request 1: OK
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Request 2: OK
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Request 3: Rate Limited (429)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)

	// Wait for window reset
	time.Sleep(60 * time.Millisecond)

	// Request 4: OK again after reset
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// Test default RateLimiter invocation
	rDefault := gin.New()
	rDefault.Use(RateLimiter())
	rDefault.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	wDef := httptest.NewRecorder()
	reqDef, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	rDefault.ServeHTTP(wDef, reqDef)
	assert.Equal(t, http.StatusOK, wDef.Code)
}
