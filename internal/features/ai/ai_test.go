package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newMockHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestSearchClient(t *testing.T) {
	// 1. Default constructor
	cDef := NewSearchClient("https://delcom.org/api/v1/search", "key")
	assert.NotNil(t, cDef)

	// 2. Empty endpoint
	cEmpty := NewSearchClientWithHTTPClient("", "", http.DefaultClient)
	resEmpty, err := cEmpty.Search("golang")
	assert.NoError(t, err)
	assert.Empty(t, resEmpty)

	// 3. Invalid URL
	cInvalid := NewSearchClientWithHTTPClient("::invalid url", "", http.DefaultClient)
	_, err = cInvalid.Search("golang")
	assert.Error(t, err)

	// 4. Success search with auth header
	mockClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "Bearer testkey", req.Header.Get("Authorization"))
		assert.Equal(t, "golang", req.URL.Query().Get("q"))
		resp := SearchEngineResponse{
			Success: true,
			Message: "OK",
			Data: []SearchSource{
				{Title: "Go Lang", URL: "https://golang.org", Snippet: "Go is an open source programming language."},
			},
		}
		data, _ := json.Marshal(resp)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
			Header:     make(http.Header),
		}, nil
	})

	cSuccess := NewSearchClientWithHTTPClient("https://search.delcom.org/search", "testkey", mockClient)
	res, err := cSuccess.Search("golang")
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "Go Lang", res[0].Title)

	// 5. Server error (non-200)
	mockErrClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewReader([]byte("server error"))),
			Header:     make(http.Header),
		}, nil
	})
	cErr := NewSearchClientWithHTTPClient("https://search.delcom.org/search", "", mockErrClient)
	_, err = cErr.Search("golang")
	assert.Error(t, err)

	// 6. Invalid JSON response
	mockBadJSONClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("not json"))),
			Header:     make(http.Header),
		}, nil
	})
	cBadJSON := NewSearchClientWithHTTPClient("https://search.delcom.org/search", "", mockBadJSONClient)
	_, err = cBadJSON.Search("golang")
	assert.Error(t, err)

	// 7. Network/Client transport failure
	mockFailClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network timeout")
	})
	cFail := NewSearchClientWithHTTPClient("https://search.delcom.org/search", "", mockFailClient)
	_, err = cFail.Search("golang")
	assert.Error(t, err)
}

func TestOpenAIClient(t *testing.T) {
	// Test constructor default model & endpoint suffix
	cliDef := NewOpenAIClient("https://api.openai.com/v1", "key", "")
	assert.NotNil(t, cliDef)

	// 1. Success chat
	mockClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "Bearer aikey", req.Header.Get("Authorization"))
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		resp := OpenAIChatResponse{
			Choices: []OpenAIChatChoice{
				{
					Index: 0,
					Message: OpenAIChatMessage{
						Role:    "assistant",
						Content: "Konten AI",
					},
				},
			},
		}
		data, _ := json.Marshal(resp)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewOpenAIClientWithHTTPClient("https://delcom.org/api/v1/chat/completions", "aikey", "gpt-4o-mini", mockClient)
	content, err := client.Chat("sys", "user")
	assert.NoError(t, err)
	assert.Equal(t, "Konten AI", content)

	// 2. Non-200 error
	mockErrClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewReader([]byte("bad request"))),
			Header:     make(http.Header),
		}, nil
	})
	clientErr := NewOpenAIClientWithHTTPClient("https://delcom.org/api/v1", "", "gpt-4o-mini", mockErrClient)
	_, err = clientErr.Chat("sys", "user")
	assert.Error(t, err)

	// 3. Invalid JSON
	mockBadJSONClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("{invalid"))),
			Header:     make(http.Header),
		}, nil
	})
	clientBad := NewOpenAIClientWithHTTPClient("https://delcom.org/api/v1", "", "gpt-4o-mini", mockBadJSONClient)
	_, err = clientBad.Chat("sys", "user")
	assert.Error(t, err)

	// 4. Empty choices
	mockEmptyChoicesClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		resp := OpenAIChatResponse{Choices: []OpenAIChatChoice{}}
		data, _ := json.Marshal(resp)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
			Header:     make(http.Header),
		}, nil
	})
	clientEmpty := NewOpenAIClientWithHTTPClient("https://delcom.org/api/v1", "", "gpt-4o-mini", mockEmptyChoicesClient)
	_, err = clientEmpty.Chat("sys", "user")
	assert.Error(t, err)

	// 5. Network failure
	mockFailClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	})
	clientDead := NewOpenAIClientWithHTTPClient("https://delcom.org/api/v1", "", "gpt-4o-mini", mockFailClient)
	_, err = clientDead.Chat("sys", "user")
	assert.Error(t, err)

	// 6. Invalid endpoint URL error in NewRequest
	clientBadURL := NewOpenAIClientWithHTTPClient("::invalid url", "", "gpt-4o-mini", mockFailClient)
	_, err = clientBadURL.Chat("sys", "user")
	assert.Error(t, err)
}

type mockOpenAI struct {
	response string
	err      error
}

func (m *mockOpenAI) Chat(sys, user string) (string, error) {
	return m.response, m.err
}

type mockSearch struct {
	sources []SearchSource
	err     error
}

func (m *mockSearch) Search(query string) ([]SearchSource, error) {
	return m.sources, m.err
}

func TestAIService(t *testing.T) {
	// 1. Success with Web Search & fenced json
	mockAI := &mockOpenAI{
		response: "```json\n{\n  \"title\": \"Judul Golang\",\n  \"category\": \"Teknologi\",\n  \"excerpt\": \"Ringkasan\",\n  \"content\": \"Konten markdown lengkap\"\n}\n```",
	}
	mockS := &mockSearch{
		sources: []SearchSource{
			{Title: "Golang News", URL: "https://news.com", Snippet: "Latest updates"},
		},
	}
	svc := NewService(mockAI, mockS)
	req := &AiGenerateRequest{
		Topic:            "Golang Web",
		Category:         "Teknologi",
		IncludeWebSearch: true,
	}

	res, err := svc.GenerateBlog(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "Judul Golang", res.Title)
	assert.Equal(t, "Teknologi", res.Category)
	assert.Len(t, res.Sources, 1)

	// 2. Success with empty category and search error (fallback to no sources)
	mockS.err = errors.New("search failed")
	mockAI.response = `{"title": "", "category": "", "excerpt": "Ringkas", "content": "Konten"}`
	req2 := &AiGenerateRequest{
		Topic:            "AI Trends",
		Category:         "",
		IncludeWebSearch: true,
	}
	res2, err := svc.GenerateBlog(context.Background(), req2)
	assert.NoError(t, err)
	assert.Equal(t, "AI Trends", res2.Title)
	assert.Equal(t, "Umum", res2.Category)
	assert.Empty(t, res2.Sources)

	// 3. Fallback non-JSON raw response
	mockAI.response = "Ini bukan format JSON yang valid melainkan teks biasa."
	req3 := &AiGenerateRequest{
		Topic:            "Teks Biasa",
		Category:         "Pendidikan",
		IncludeWebSearch: false,
	}
	res3, err := svc.GenerateBlog(context.Background(), req3)
	assert.NoError(t, err)
	assert.Equal(t, "Teks Biasa", res3.Title)
	assert.Equal(t, "Pendidikan", res3.Category)
	assert.Equal(t, "Ini bukan format JSON yang valid melainkan teks biasa.", res3.Content)

	// 4. OpenAI Chat error
	mockAI.err = errors.New("rate limited")
	_, err = svc.GenerateBlog(context.Background(), req)
	assert.Error(t, err)
}

func TestAIHandler(t *testing.T) {
	mockAI := &mockOpenAI{
		response: `{"title": "Test Judul", "category": "Umum", "excerpt": "Excerpt", "content": "Content"}`,
	}
	svc := NewService(mockAI, nil)
	h := NewHandler(svc)

	r := gin.New()
	r.POST("/api/posts/ai-generate", h.GenerateArticle)

	// 1. Success
	body, _ := json.Marshal(AiGenerateRequest{Topic: "Tema Baru"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/posts/ai-generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Empty topic
	bodyEmpty, _ := json.Marshal(AiGenerateRequest{Topic: ""})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/posts/ai-generate", bytes.NewBuffer(bodyEmpty))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 3. Invalid JSON payload
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/posts/ai-generate", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 4. Service failure (500)
	mockAI.err = errors.New("chat failed")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/posts/ai-generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
