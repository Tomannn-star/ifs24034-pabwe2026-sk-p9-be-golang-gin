package post

import (
	"strings"

	"app/internal/database"

	"gorm.io/gorm"
)

type Repository interface {
	FindAllPublic(q, category string) ([]database.Post, error)
	FindBySlugPublic(slug string) (*database.Post, error)
	FindAllByUserID(userID uint) ([]database.Post, error)
	FindByIDAndUserID(id, userID uint) (*database.Post, error)
	FindByID(id uint) (*database.Post, error)
	FindBySlug(slug string) (*database.Post, error)
	Create(post *database.Post) error
	Update(post *database.Post) error
	Delete(post *database.Post) error
	GetStatsByUserID(userID uint) (*PostStatsResponse, error)
	FindAllAdminPosts() ([]database.Post, error)
	FindAllAdminUsers() ([]AdminUserStatsResponse, error)
	FindAllByUserIDAdmin(userID uint) ([]database.Post, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAllPublic(q, category string) ([]database.Post, error) {
	var posts []database.Post
	query := r.db.Preload("User").Where("is_published = ?", true)

	if category != "" && category != "Semua" {
		query = query.Where("category = ?", category)
	}

	if q != "" {
		search := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(excerpt) LIKE ? OR LOWER(content) LIKE ?", search, search, search)
	}

	err := query.Order("published_at desc, id desc").Find(&posts).Error
	return posts, err
}

func (r *repository) FindBySlugPublic(slug string) (*database.Post, error) {
	var post database.Post
	if err := r.db.Preload("User").Where("slug = ? AND is_published = ?", slug, true).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repository) FindAllByUserID(userID uint) ([]database.Post, error) {
	var posts []database.Post
	err := r.db.Preload("User").Where("user_id = ?", userID).Order("id desc").Find(&posts).Error
	return posts, err
}

func (r *repository) FindByIDAndUserID(id, userID uint) (*database.Post, error) {
	var post database.Post
	if err := r.db.Preload("User").Where("id = ? AND user_id = ?", id, userID).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repository) FindByID(id uint) (*database.Post, error) {
	var post database.Post
	err := r.db.Preload("User").Where("id = ?", id).First(&post).Error
	return &post, err
}

func (r *repository) FindBySlug(slug string) (*database.Post, error) {
	var post database.Post
	if err := r.db.Where("slug = ?", slug).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repository) Create(post *database.Post) error {
	return r.db.Create(post).Error
}

func (r *repository) Update(post *database.Post) error {
	return r.db.Save(post).Error
}

func (r *repository) Delete(post *database.Post) error {
	return r.db.Delete(post).Error
}

func (r *repository) GetStatsByUserID(userID uint) (*PostStatsResponse, error) {
	var total, published int64
	_ = r.db.Model(&database.Post{}).Where("user_id = ?", userID).Count(&total).Error
	_ = r.db.Model(&database.Post{}).Where("user_id = ? AND is_published = ?", userID, true).Count(&published).Error
	draft := total - published
	return &PostStatsResponse{
		Total:          total,
		TotalArticles:  total,
		TotalPosts:     total,
		TotalPublished: published,
		TotalTerbit:    published,
		TotalSelesai:   published,
		TotalDraft:     draft,
		BelumSelesai:   draft,
	}, nil
}

func (r *repository) FindAllAdminPosts() ([]database.Post, error) {
	var posts []database.Post
	err := r.db.Preload("User").Order("id desc").Find(&posts).Error
	return posts, err
}

func (r *repository) FindAllAdminUsers() ([]AdminUserStatsResponse, error) {
	var users []database.User
	_ = r.db.Preload("Posts").Order("id asc").Find(&users).Error

	result := []AdminUserStatsResponse{}
	for _, u := range users {
		var totalPosts, publishedPosts int64
		for _, p := range u.Posts {
			totalPosts++
			if p.IsPublished {
				publishedPosts++
			}
		}
		draftPosts := totalPosts - publishedPosts
		result = append(result, AdminUserStatsResponse{
			ID:                u.ID,
			Username:          u.Username,
			Name:              u.Username,
			Email:             u.Email,
			Role:              u.Role,
			Total:             totalPosts,
			TotalPosts:        totalPosts,
			TotalArticles:     totalPosts,
			TotalTodos:        totalPosts,
			PublishedPosts:    publishedPosts,
			PublishedArticles: publishedPosts,
			TotalPublished:    publishedPosts,
			TotalTerbit:       publishedPosts,
			TotalSelesai:      publishedPosts,
			CompletedPosts:    publishedPosts,
			CompletedTodos:    publishedPosts,
			TotalDraft:        draftPosts,
			DraftPosts:        draftPosts,
			BelumSelesai:      draftPosts,
			CreatedAt:         u.CreatedAt,
		})
	}
	return result, nil
}

func (r *repository) FindAllByUserIDAdmin(userID uint) ([]database.Post, error) {
	var posts []database.Post
	err := r.db.Preload("User").Where("user_id = ?", userID).Order("id desc").Find(&posts).Error
	return posts, err
}
