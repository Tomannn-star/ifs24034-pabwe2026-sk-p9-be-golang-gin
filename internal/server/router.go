package server

import (
	"app/docs"
	"app/internal/config"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/middleware"
	"app/internal/response"
	"crypto/tls"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	cfg *config.Config,
	authH *auth.Handler,
	postH *post.Handler,
	aiH *ai.Handler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter())

	// 404 NoRoute handler
	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, http.StatusNotFound, "Endpoint not found")
	})

	// Swagger & OpenAPI JSON
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/api.json", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, docs.GetSwaggerJSON(c.Request.Host, scheme))
	})

	api := r.Group("/api")
	{
		// Public routes
		public := api.Group("/public")
		{
			public.GET("/categories", postH.GetCategories)
			public.GET("/posts", postH.GetPublicPosts)
			public.GET("/posts/:slug", postH.GetPublicPostBySlug)
		}

		// Auth routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authH.Register)
			authGroup.POST("/login", authH.Login)

			authProtected := authGroup.Group("")
			authProtected.Use(middleware.Auth(cfg.AppKey))
			{
				authProtected.POST("/logout", authH.Logout)
				authProtected.GET("/profile", authH.GetProfile)
				authProtected.PUT("/profile", authH.UpdateProfile)
			}
		}

		// Author Post routes (Authenticated)
		postsGroup := api.Group("/posts")
		postsGroup.Use(middleware.Auth(cfg.AppKey))
		{
			postsGroup.GET("", postH.GetAuthorPosts)
			postsGroup.POST("", postH.CreatePost)
			postsGroup.GET("/stats", postH.GetPostStats)
			postsGroup.GET("/:id", postH.GetAuthorPostByID)
			postsGroup.PUT("/:id", postH.UpdateAuthorPost)
			postsGroup.PATCH("/:id/toggle-publish", postH.TogglePublish)
			postsGroup.DELETE("/:id", postH.DeleteAuthorPost)
			postsGroup.POST("/ai-generate", aiH.GenerateArticle)
		}

		// Admin routes (Authenticated + RequireRole("admin"))
		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.Auth(cfg.AppKey), middleware.RequireRole("admin"))
		{
			adminGroup.GET("/users", postH.GetAdminUsers)
			adminGroup.GET("/posts", postH.GetAdminPosts)
			adminGroup.GET("/users/:id/posts", postH.GetAdminUserPosts)
		}
	}

	return r
}

func init() {
	// Reference to ensure tls import isn't unused
	_ = (*tls.ConnectionState)(nil)
}
