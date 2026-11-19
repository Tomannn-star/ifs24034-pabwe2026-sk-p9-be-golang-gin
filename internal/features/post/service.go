package post

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"app/internal/database"
)

var StandardCategories = []string{"Teknologi", "Pendidikan", "Gaya Hidup", "Bisnis", "Opini", "Umum"}

var (
	ErrPostNotFound = errors.New("post not found")
	ErrForbidden    = errors.New("forbidden")
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

func CalculateReadingTime(content string) int {
	words := strings.Fields(content)
	count := len(words)
	if count == 0 {
		return 1
	}
	return (count + 199) / 200
}

func GenerateSlug(title string) string {
	lower := strings.ToLower(title)
	clean := nonAlphanumericRegex.ReplaceAllString(lower, "-")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		return "post"
	}
	return clean
}

type Service interface {
	GetCategories() []string
	GetPublicPosts(q, category string) ([]PostResponse, error)
	GetPublicPostBySlug(slug string) (*PostResponse, error)
	GetAuthorPosts(userID uint) ([]PostResponse, error)
	CreatePost(userID uint, req *CreatePostRequest) (*PostResponse, error)
	GetPostStats(userID uint) (*PostStatsResponse, error)
	GetAuthorPostByID(id, userID uint) (*PostResponse, error)
	UpdateAuthorPost(id, userID uint, req *UpdatePostRequest) (*PostResponse, error)
	TogglePublish(id, userID uint) (*PostResponse, error)
	DeleteAuthorPost(id, userID uint) error
	GetAdminUsers() ([]AdminUserStatsResponse, error)
	GetAdminPosts() ([]PostResponse, error)
	GetAdminUserPosts(userID uint) ([]PostResponse, error)
	GenerateUniqueSlug(title string, excludeID uint) string
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetCategories() []string {
	return StandardCategories
}

func (s *service) GenerateUniqueSlug(title string, excludeID uint) string {
	baseSlug := GenerateSlug(title)
	slug := baseSlug
	counter := 1

	for {
		existing, err := s.repo.FindBySlug(slug)
		if err != nil || existing == nil || (excludeID > 0 && existing.ID == excludeID) {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug
}

func toPostResponse(p *database.Post) PostResponse {
	authorName := ""
	if p.User != nil {
		authorName = p.User.Username
	}
	return PostResponse{
		ID:             p.ID,
		UserID:         p.UserID,
		AuthorName:     authorName,
		Title:          p.Title,
		Slug:           p.Slug,
		Category:       p.Category,
		Excerpt:        p.Excerpt,
		Content:        p.Content,
		CoverImagePath: p.CoverImagePath,
		ReadingTime:    CalculateReadingTime(p.Content),
		IsPublished:    p.IsPublished,
		PublishedAt:    p.PublishedAt,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func toPublicPostResponse(p *database.Post) PostResponse {
	res := toPostResponse(p)
	res.UserID = 0
	return res
}

func (s *service) GetPublicPosts(q, category string) ([]PostResponse, error) {
	posts, err := s.repo.FindAllPublic(q, category)
	res := []PostResponse{}
	for _, p := range posts {
		res = append(res, toPublicPostResponse(&p))
	}
	return res, err
}

func (s *service) GetPublicPostBySlug(slug string) (*PostResponse, error) {
	post, err := s.repo.FindBySlugPublic(slug)
	if err != nil {
		return nil, errors.New("post not found")
	}
	res := toPublicPostResponse(post)
	return &res, nil
}

func (s *service) GetAuthorPosts(userID uint) ([]PostResponse, error) {
	posts, err := s.repo.FindAllByUserID(userID)
	res := []PostResponse{}
	for _, p := range posts {
		res = append(res, toPostResponse(&p))
	}
	return res, err
}

func (s *service) CreatePost(userID uint, req *CreatePostRequest) (*PostResponse, error) {
	slug := s.GenerateUniqueSlug(req.Title, 0)
	isPublished := false
	var publishedAt *time.Time
	if req.IsPublished != nil && *req.IsPublished {
		isPublished = true
		now := time.Now()
		publishedAt = &now
	}

	category := req.Category
	if category == "" {
		category = "Umum"
	}

	post := &database.Post{
		UserID:         userID,
		Title:          req.Title,
		Slug:           slug,
		Category:       category,
		Excerpt:        req.Excerpt,
		Content:        req.Content,
		CoverImagePath: req.CoverImagePath,
		IsPublished:    isPublished,
		PublishedAt:    publishedAt,
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	saved, _ := s.repo.FindByID(post.ID)
	res := toPostResponse(saved)
	return &res, nil
}

func (s *service) GetPostStats(userID uint) (*PostStatsResponse, error) {
	return s.repo.GetStatsByUserID(userID)
}

func (s *service) GetAuthorPostByID(id, userID uint) (*PostResponse, error) {
	post, err := s.repo.FindByID(id)
	if err != nil || post == nil {
		return nil, ErrPostNotFound
	}
	if post.UserID != userID {
		return nil, ErrForbidden
	}
	res := toPostResponse(post)
	return &res, nil
}

func (s *service) UpdateAuthorPost(id, userID uint, req *UpdatePostRequest) (*PostResponse, error) {
	post, err := s.repo.FindByID(id)
	if err != nil || post == nil {
		return nil, ErrPostNotFound
	}
	if post.UserID != userID {
		return nil, ErrForbidden
	}

	if req.Title != "" && req.Title != post.Title {
		post.Title = req.Title
		post.Slug = s.GenerateUniqueSlug(req.Title, post.ID)
	}

	if req.Category != "" {
		post.Category = req.Category
	}

	if req.Excerpt != "" {
		post.Excerpt = req.Excerpt
	}

	if req.Content != "" {
		post.Content = req.Content
	}

	if req.CoverImagePath != "" {
		post.CoverImagePath = req.CoverImagePath
	}

	if req.IsPublished != nil {
		if *req.IsPublished && !post.IsPublished {
			post.IsPublished = true
			now := time.Now()
			post.PublishedAt = &now
		} else if !*req.IsPublished && post.IsPublished {
			post.IsPublished = false
			post.PublishedAt = nil
		}
	}

	if err := s.repo.Update(post); err != nil {
		return nil, err
	}

	saved, _ := s.repo.FindByID(post.ID)
	res := toPostResponse(saved)
	return &res, nil
}

func (s *service) TogglePublish(id, userID uint) (*PostResponse, error) {
	post, err := s.repo.FindByID(id)
	if err != nil || post == nil {
		return nil, ErrPostNotFound
	}
	if post.UserID != userID {
		return nil, ErrForbidden
	}

	if post.IsPublished {
		post.IsPublished = false
		post.PublishedAt = nil
	} else {
		post.IsPublished = true
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := s.repo.Update(post); err != nil {
		return nil, err
	}

	saved, _ := s.repo.FindByID(post.ID)
	res := toPostResponse(saved)
	return &res, nil
}

func (s *service) DeleteAuthorPost(id, userID uint) error {
	post, err := s.repo.FindByID(id)
	if err != nil || post == nil {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(post)
}

func (s *service) GetAdminUsers() ([]AdminUserStatsResponse, error) {
	return s.repo.FindAllAdminUsers()
}

func (s *service) GetAdminPosts() ([]PostResponse, error) {
	posts, err := s.repo.FindAllAdminPosts()
	res := []PostResponse{}
	for _, p := range posts {
		res = append(res, toPostResponse(&p))
	}
	return res, err
}

func (s *service) GetAdminUserPosts(userID uint) ([]PostResponse, error) {
	posts, err := s.repo.FindAllByUserIDAdmin(userID)
	res := []PostResponse{}
	for _, p := range posts {
		res = append(res, toPostResponse(&p))
	}
	return res, err
}
