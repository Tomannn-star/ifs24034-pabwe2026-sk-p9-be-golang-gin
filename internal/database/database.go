package database

import (
	"fmt"
	"time"

	"app/internal/config"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"size:255;unique;not null" json:"username"`
	Email     string    `gorm:"size:255;unique;not null" json:"email"`
	Password  string    `gorm:"type:text;not null" json:"-"`
	Role      string    `gorm:"size:50;not null;default:'member'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Posts     []Post    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"posts,omitempty"`
}

type Post struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	User           *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title          string     `gorm:"size:255;not null" json:"title"`
	Slug           string     `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Category       string     `gorm:"size:100;not null;default:'Umum'" json:"category"`
	Excerpt        string     `gorm:"type:text" json:"excerpt"`
	Content        string     `gorm:"type:text;not null" json:"content"`
	CoverImagePath string     `gorm:"type:text" json:"cover_image_path"`
	IsPublished    bool       `gorm:"default:false" json:"is_published"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if cfg.DBConnection == "sqlite" {
		dbPath := cfg.DBDatabase
		if dbPath == "" || dbPath == ":memory:" {
			dbPath = ":memory:?cache=shared"
		}
		dialector = sqlite.Open(dbPath)
	} else {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s connect_timeout=5",
			cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBDatabase, cfg.DBPort, cfg.DBSSLMode)
		dialector = postgres.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if cfg.DBConnection == "sqlite" {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.SetMaxOpenConns(1)
			sqlDB.SetMaxIdleConns(1)
		}
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Post{})
}

func Seed(db *gorm.DB, cfg *config.Config) error {
	// Seed Admin
	var count int64
	db.Model(&User{}).Where("username = ?", cfg.AdminUsername).Count(&count)
	if count == 0 {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		admin := User{
			Username: cfg.AdminUsername,
			Email:    "admin@delcom.org",
			Password: string(hashedPass),
			Role:     "admin",
		}
		_ = db.Create(&admin).Error
	}

	// Seed Member
	var memberCount int64
	db.Model(&User{}).Where("username = ?", "member").Count(&memberCount)
	if memberCount == 0 {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("member"), bcrypt.DefaultCost)
		member := User{
			Username: "member",
			Email:    "member@delcom.org",
			Password: string(hashedPass),
			Role:     "member",
		}
		_ = db.Create(&member).Error
	}

	// Seed Initial Posts for Public & Admin tests
	var postCount int64
	db.Model(&Post{}).Count(&postCount)
	if postCount == 0 {
		var adminUser User
		if err := db.Where("username = ?", cfg.AdminUsername).First(&adminUser).Error; err == nil {
			now := time.Now()
			p1 := Post{
				UserID:      adminUser.ID,
				Title:       "Panduan Arsitektur Golang Gin dan React",
				Slug:        "panduan-arsitektur-golang-gin-dan-react",
				Category:    "Teknologi",
				Excerpt:     "Panduan arsitektur modern membangun REST API Golang Gin dan frontend React.",
				Content:     "Dalam pengembangan aplikasi web modern, arsitektur REST API dengan Golang Gin memberikan performa tinggi dan konkurensi yang optimal.",
				IsPublished: true,
				PublishedAt: &now,
			}
			p2 := Post{
				UserID:      adminUser.ID,
				Title:       "Inovasi Pembelajaran Digital di Kampus",
				Slug:        "inovasi-pembelajaran-digital-di-kampus",
				Category:    "Pendidikan",
				Excerpt:     "Bagaimana teknologi digital mentransformasi metode pembelajaran di perguruan tinggi.",
				Content:     "Pendidikan modern menuntut integrasi alat digital untuk mendukung pembelajaran kolaboratif dan adaptif.",
				IsPublished: true,
				PublishedAt: &now,
			}
			_ = db.Create(&p1).Error
			_ = db.Create(&p2).Error
		}
	}

	return nil
}
