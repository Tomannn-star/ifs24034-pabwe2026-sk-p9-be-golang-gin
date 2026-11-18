package auth

import (
	"app/internal/database"

	"gorm.io/gorm"
)

type Repository interface {
	FindByUsername(username string) (*database.User, error)
	FindByEmail(email string) (*database.User, error)
	FindByID(id uint) (*database.User, error)
	Create(user *database.User) error
	Update(user *database.User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByUsername(username string) (*database.User, error) {
	var user database.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByEmail(email string) (*database.User, error) {
	var user database.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(id uint) (*database.User, error) {
	var user database.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Create(user *database.User) error {
	return r.db.Create(user).Error
}

func (r *repository) Update(user *database.User) error {
	return r.db.Save(user).Error
}
