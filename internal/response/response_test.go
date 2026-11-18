package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponse_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusOK, "Berhasil", gin.H{"id": 1})
	assert.Equal(t, http.StatusOK, w.Code)

	var res Response
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "success", res.Status)
	assert.Equal(t, "Berhasil", res.Message)
	assert.NotNil(t, res.Data)

	// Test without data
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	Success(c2, http.StatusOK, "Berhasil tanpa data")
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestResponse_Fail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Fail(c, http.StatusBadRequest, "Gagal validasi", gin.H{"field": "required"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res Response
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "fail", res.Status)
	assert.Equal(t, "Gagal validasi", res.Message)

	// Test without data
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	Fail(c2, http.StatusBadRequest, "Gagal")
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestResponse_Error(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, http.StatusInternalServerError, "Server error", "detail error")
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var res Response
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "error", res.Status)
	assert.Equal(t, "Server error", res.Message)

	// Test without data
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	Error(c2, http.StatusInternalServerError, "Server error")
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}

func TestResponse_AbortFail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	AbortFail(c, http.StatusUnauthorized, "Unauthorized", "token expired")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())

	// Test without data
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	AbortFail(c2, http.StatusUnauthorized, "Unauthorized")
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
	assert.True(t, c2.IsAborted())
}

func TestResponse_AbortError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	AbortError(c, http.StatusInternalServerError, "Crash", "panic info")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, c.IsAborted())

	// Test without data
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	AbortError(c2, http.StatusInternalServerError, "Crash")
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
	assert.True(t, c2.IsAborted())
}
