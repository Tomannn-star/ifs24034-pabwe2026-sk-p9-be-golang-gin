package ai

import (
	"app/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GenerateArticle(c *gin.Context) {
	var req AiGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Topic is required: "+err.Error())
		return
	}

	if req.Topic == "" {
		response.Fail(c, http.StatusBadRequest, "Topic is required")
		return
	}

	res, err := h.service.GenerateBlog(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate blog content: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "AI content generated successfully", res)
}
