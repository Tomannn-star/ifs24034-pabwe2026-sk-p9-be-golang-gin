package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type SearchClient interface {
	Search(query string) ([]SearchSource, error)
}

type searchClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

func NewSearchClient(endpoint, apiKey string) SearchClient {
	return NewSearchClientWithHTTPClient(endpoint, apiKey, &http.Client{Timeout: 10 * time.Second})
}

func NewSearchClientWithHTTPClient(endpoint, apiKey string, httpClient *http.Client) SearchClient {
	return &searchClient{
		endpoint:   endpoint,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func (c *searchClient) Search(query string) ([]SearchSource, error) {
	if c.endpoint == "" {
		return []SearchSource{}, nil
	}

	reqURL, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Set("q", query)
	reqURL.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search engine returned status: %d", resp.StatusCode)
	}

	var searchResp SearchEngineResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return searchResp.Data, nil
}
