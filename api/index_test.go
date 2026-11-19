package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_PublicCategories(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/public/categories", nil)
	w := httptest.NewRecorder()

	Handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Umum")
}

func TestHandler_NotFound(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	w := httptest.NewRecorder()

	Handler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
