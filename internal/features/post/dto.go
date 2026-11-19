package post

import "time"

type CreatePostRequest struct {
	Title          string `json:"title" binding:"required"`
	Category       string `json:"category"`
	Excerpt        string `json:"excerpt"`
	Content        string `json:"content" binding:"required"`
	CoverImagePath string `json:"cover_image_path"`
	IsPublished    *bool  `json:"is_published"`
}

type UpdatePostRequest struct {
	Title          string `json:"title"`
	Category       string `json:"category"`
	Excerpt        string `json:"excerpt"`
	Content        string `json:"content"`
	CoverImagePath string `json:"cover_image_path"`
	IsPublished    *bool  `json:"is_published"`
}

type PostResponse struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"user_id,omitempty"`
	AuthorName     string     `json:"author_name"`
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Category       string     `json:"category"`
	Excerpt        string     `json:"excerpt"`
	Content        string     `json:"content"`
	CoverImagePath string     `json:"cover_image_path"`
	ReadingTime    int        `json:"reading_time"`
	IsPublished    bool       `json:"is_published"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type PostStatsResponse struct {
	Total          int64 `json:"total"`
	TotalArticles  int64 `json:"total_articles"`
	TotalPosts     int64 `json:"total_posts"`
	TotalPublished int64 `json:"total_published"`
	TotalTerbit    int64 `json:"total_terbit"`
	TotalSelesai   int64 `json:"total_selesai"`
	TotalDraft     int64 `json:"total_draft"`
	BelumSelesai   int64 `json:"belum_selesai"`
}

type AdminUserStatsResponse struct {
	ID                uint      `json:"id"`
	Username          string    `json:"username"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Role              string    `json:"role"`
	Total             int64     `json:"total"`
	TotalPosts        int64     `json:"total_posts"`
	TotalArticles     int64     `json:"total_articles"`
	TotalTodos        int64     `json:"total_todos"`
	PublishedPosts    int64     `json:"published_posts"`
	PublishedArticles int64     `json:"published_articles"`
	TotalPublished    int64     `json:"total_published"`
	TotalTerbit       int64     `json:"total_terbit"`
	TotalSelesai      int64     `json:"total_selesai"`
	CompletedPosts    int64     `json:"completed_posts"`
	CompletedTodos    int64     `json:"completed_todos"`
	TotalDraft        int64     `json:"total_draft"`
	DraftPosts        int64     `json:"draft_posts"`
	BelumSelesai      int64     `json:"belum_selesai"`
	CreatedAt         time.Time `json:"created_at"`
}
