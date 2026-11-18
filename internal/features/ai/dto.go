package ai

type AiGenerateRequest struct {
	Topic            string `json:"topic"`
	Category         string `json:"category"`
	IncludeWebSearch bool   `json:"include_web_search"`
}

type SearchSource struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type AiGenerateResponse struct {
	Title    string         `json:"title"`
	Category string         `json:"category"`
	Excerpt  string         `json:"excerpt"`
	Content  string         `json:"content"`
	Sources  []SearchSource `json:"sources"`
}

type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
}

type OpenAIChatChoice struct {
	Index   int               `json:"index"`
	Message OpenAIChatMessage `json:"message"`
}

type OpenAIChatResponse struct {
	Choices []OpenAIChatChoice `json:"choices"`
}

type SearchEngineResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    []SearchSource `json:"data"`
}
