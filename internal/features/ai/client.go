package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient interface {
	Chat(systemPrompt, userPrompt string) (string, error)
}

type openAIClient struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAIClient(endpoint, apiKey, model string) OpenAIClient {
	return NewOpenAIClientWithHTTPClient(endpoint, apiKey, model, &http.Client{Timeout: 30 * time.Second})
}

func NewOpenAIClientWithHTTPClient(endpoint, apiKey, model string, httpClient *http.Client) OpenAIClient {
	ep := strings.TrimRight(endpoint, "/")
	if !strings.HasSuffix(ep, "/chat/completions") {
		ep = ep + "/chat/completions"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &openAIClient{
		endpoint:   ep,
		apiKey:     apiKey,
		model:      model,
		httpClient: httpClient,
	}
}

func (c *openAIClient) Chat(systemPrompt, userPrompt string) (string, error) {
	reqBody := OpenAIChatRequest{
		Model: c.model,
		Messages: []OpenAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai returned status: %d", resp.StatusCode)
	}

	var aiResp OpenAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return "", err
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from openai")
	}

	return aiResp.Choices[0].Message.Content, nil
}
