package post

import "time"

type CreatePostRequest struct {
	Title          string `json:"title" binding:"required"`
	Category       string `json:"category" binding:"required"`
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
	UserID         uint       `json:"user_id"`
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
	TotalPosts     int64 `json:"total_posts"`
	TotalPublished int64 `json:"total_published"`
	TotalTerbit    int64 `json:"total_terbit"`
	TotalDraft     int64 `json:"total_draft"`
	BelumSelesai   int64 `json:"belum_selesai"`
}

type AdminUserStatsResponse struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	TotalPosts     int64     `json:"total_posts"`
	PublishedPosts int64     `json:"published_posts"`
	CreatedAt      time.Time `json:"created_at"`
}
