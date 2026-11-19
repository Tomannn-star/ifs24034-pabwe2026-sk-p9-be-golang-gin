package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_Minimal(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()

	Handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "minimal handler works")
}
