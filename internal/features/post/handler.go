package post

import (
	"errors"
	"net/http"
	"strconv"

	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func getUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := val.(uint)
	return id, ok
}

func (h *Handler) GetCategories(c *gin.Context) {
	categories := h.service.GetCategories()
	response.Success(c, http.StatusOK, "Categories retrieved successfully", categories)
}

func (h *Handler) GetPublicPosts(c *gin.Context) {
	q := c.Query("q")
	category := c.Query("category")

	posts, err := h.service.GetPublicPosts(q, category)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}
	response.Success(c, http.StatusOK, "Public posts retrieved successfully", posts)
}

func (h *Handler) GetPublicPostBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := h.service.GetPublicPostBySlug(slug)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "Post not found")
		return
	}
	response.Success(c, http.StatusOK, "Post retrieved successfully", post)
}

func (h *Handler) GetAuthorPosts(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	posts, err := h.service.GetAuthorPosts(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve author posts")
		return
	}
	response.Success(c, http.StatusOK, "Author posts retrieved successfully", posts)
}

func (h *Handler) CreatePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	post, err := h.service.CreatePost(userID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create post")
		return
	}
	response.Success(c, http.StatusCreated, "Post created successfully", post)
}

func (h *Handler) GetPostStats(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.service.GetPostStats(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve post stats")
		return
	}
	response.Success(c, http.StatusOK, "Post stats retrieved successfully", stats)
}

func (h *Handler) GetAuthorPostByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := h.service.GetAuthorPostByID(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.Fail(c, http.StatusForbidden, "Forbidden: you do not have permission to access this post")
			return
		}
		response.Fail(c, http.StatusNotFound, "Post not found")
		return
	}
	response.Success(c, http.StatusOK, "Post retrieved successfully", post)
}

func (h *Handler) UpdateAuthorPost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	post, err := h.service.UpdateAuthorPost(uint(id), userID, &req)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.Fail(c, http.StatusForbidden, "Forbidden: you do not have permission to access this post")
			return
		}
		if errors.Is(err, ErrPostNotFound) {
			response.Fail(c, http.StatusNotFound, "Post not found")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Post updated successfully", post)
}

func (h *Handler) TogglePublish(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := h.service.TogglePublish(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.Fail(c, http.StatusForbidden, "Forbidden: you do not have permission to access this post")
			return
		}
		if errors.Is(err, ErrPostNotFound) {
			response.Fail(c, http.StatusNotFound, "Post not found")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Post publish status toggled successfully", post)
}

func (h *Handler) DeleteAuthorPost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	if err := h.service.DeleteAuthorPost(uint(id), userID); err != nil {
		if errors.Is(err, ErrForbidden) {
			response.Fail(c, http.StatusForbidden, "Forbidden: you do not have permission to access this post")
			return
		}
		if errors.Is(err, ErrPostNotFound) {
			response.Fail(c, http.StatusNotFound, "Post not found")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Post deleted successfully")
}

func (h *Handler) GetAdminUsers(c *gin.Context) {
	users, err := h.service.GetAdminUsers()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve admin users")
		return
	}
	response.Success(c, http.StatusOK, "Admin users retrieved successfully", users)
}

func (h *Handler) GetAdminPosts(c *gin.Context) {
	posts, err := h.service.GetAdminPosts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve admin posts")
		return
	}
	response.Success(c, http.StatusOK, "Admin posts retrieved successfully", posts)
}

func (h *Handler) GetAdminUserPosts(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	posts, err := h.service.GetAdminUserPosts(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve admin user posts")
		return
	}
	response.Success(c, http.StatusOK, "Admin user posts retrieved successfully", posts)
}
