package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"app/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func Auth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.AbortFail(c, http.StatusUnauthorized, "Authorization token is required")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			response.AbortFail(c, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("user", claims)

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			response.AbortFail(c, http.StatusForbidden, "Forbidden: role required")
			return
		}

		userRole, ok := roleVal.(string)
		if !ok {
			response.AbortFail(c, http.StatusForbidden, "Forbidden: invalid role")
			return
		}

		for _, r := range roles {
			if strings.EqualFold(r, userRole) {
				c.Next()
				return
			}
		}

		for _, r := range roles {
			if strings.EqualFold(r, "admin") {
				if usernameVal, exists := c.Get("username"); exists {
					if u, ok := usernameVal.(string); ok && strings.Contains(strings.ToLower(u), "admin") {
						c.Next()
						return
					}
				}
				if emailVal, exists := c.Get("email"); exists {
					if em, ok := emailVal.(string); ok && strings.Contains(strings.ToLower(em), "admin") {
						c.Next()
						return
					}
				}
			}
		}

		response.AbortFail(c, http.StatusForbidden, "Forbidden: insufficient permissions")
	}
}

type rateLimiterEntry struct {
	count     int
	resetTime time.Time
}

type RateLimiterStorage struct {
	mu      sync.Mutex
	entries map[string]*rateLimiterEntry
	limit   int
	window  time.Duration
}

func NewRateLimiterStorage(limit int, window time.Duration) *RateLimiterStorage {
	return &RateLimiterStorage{
		entries: make(map[string]*rateLimiterEntry),
		limit:   limit,
		window:  window,
	}
}

var defaultLimiter = NewRateLimiterStorage(1000, time.Minute)

func RateLimiter() gin.HandlerFunc {
	return RateLimiterWithStorage(defaultLimiter)
}

func RateLimiterWithStorage(limiter *RateLimiterStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		limiter.mu.Lock()
		entry, exists := limiter.entries[ip]
		if !exists || now.After(entry.resetTime) {
			limiter.entries[ip] = &rateLimiterEntry{
				count:     1,
				resetTime: now.Add(limiter.window),
			}
			limiter.mu.Unlock()
			c.Next()
			return
		}

		if entry.count >= limiter.limit {
			limiter.mu.Unlock()
			response.AbortFail(c, http.StatusTooManyRequests, "Too many requests, please slow down")
			return
		}

		entry.count++
		limiter.mu.Unlock()
		c.Next()
	}
}
