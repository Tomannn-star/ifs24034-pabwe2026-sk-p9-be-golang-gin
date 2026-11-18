package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Service interface {
	GenerateBlog(ctx context.Context, req *AiGenerateRequest) (*AiGenerateResponse, error)
}

type service struct {
	openAI OpenAIClient
	search SearchClient
}

func NewService(openAI OpenAIClient, search SearchClient) Service {
	return &service{
		openAI: openAI,
		search: search,
	}
}

func (s *service) GenerateBlog(ctx context.Context, req *AiGenerateRequest) (*AiGenerateResponse, error) {
	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "Umum"
	}

	var sources []SearchSource
	webContext := ""

	if req.IncludeWebSearch && s.search != nil {
		results, err := s.search.Search(req.Topic)
		if err == nil && len(results) > 0 {
			sources = results
			var sb strings.Builder
			sb.WriteString("\n\nBerikut adalah hasil rujukan web aktual untuk topik tersebut:\n")
			for i, r := range results {
				sb.WriteString(fmt.Sprintf("%d. Judul: %s\n   URL: %s\n   Kutipan: %s\n", i+1, r.Title, r.URL, r.Snippet))
			}
			webContext = sb.String()
		}
	}

	systemPrompt := `Anda adalah asisten penulis artikel blog profesional berbahasa Indonesia yang ahli dalam SEO dan penulisan konten menarik.
Tugas Anda adalah menghasilkan draf artikel blog utuh berdasarkan topik yang diminta.
Format output WAJIB HANYA berupa JSON murni tanpa markdown code block atau teks lainnya di luar JSON:
{
  "title": "Judul Artikel yang Menarik dan Informatif",
  "category": "Kategori Artikel",
  "excerpt": "Ringkasan padat dan menarik 1-2 kalimat (maksimal 200 karakter)",
  "content": "Konten artikel lengkap dalam format Markdown kaya (menggunakan heading ##, ###, bullet points, penekanan teks tebal/miring, dan cantumkan sitasi sumber jika diberikan rujukan web)"
}`

	userPrompt := fmt.Sprintf("Topik artikel: %s\nKategori pilihan: %s%s\n\nBuatkan artikel blog berkualitas tinggi dalam format JSON sesuai instruksi.", req.Topic, category, webContext)

	rawResponse, err := s.openAI.Chat(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("openai chat error: %w", err)
	}

	// Clean code blocks if present
	cleanJSON := strings.TrimSpace(rawResponse)
	if strings.HasPrefix(cleanJSON, "```") {
		lines := strings.Split(cleanJSON, "\n")
		if len(lines) > 2 {
			cleanJSON = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	cleanJSON = strings.TrimSpace(cleanJSON)

	var genResp AiGenerateResponse
	if err := json.Unmarshal([]byte(cleanJSON), &genResp); err != nil {
		genResp = AiGenerateResponse{
			Title:    req.Topic,
			Category: category,
			Excerpt:  "Artikel tentang " + req.Topic,
			Content:  rawResponse,
		}
	}

	if genResp.Title == "" {
		genResp.Title = req.Topic
	}
	if genResp.Category == "" {
		genResp.Category = category
	}
	if sources == nil {
		sources = []SearchSource{}
	}
	genResp.Sources = sources

	return &genResp, nil
}
