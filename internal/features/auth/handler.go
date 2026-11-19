package auth

import (
	"net/http"

	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	res, err := h.service.Register(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	identifier := req.Username
	if identifier == "" {
		identifier = req.Email
	}
	if identifier == "" {
		identifier = req.Identity
	}

	if identifier == "" || req.Password == "" {
		response.Fail(c, http.StatusBadRequest, "Username/email and password are required")
		return
	}

	if req.Username == "" {
		req.Username = identifier
	}

	res, err := h.service.Login(&req)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", res)
}

func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, "Logout successful")
}

func (h *Handler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Invalid user session")
		return
	}

	user, err := h.service.GetProfile(userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", user)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "Invalid user session")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	res, err := h.service.UpdateProfile(userID, &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile updated successfully", res)
}
