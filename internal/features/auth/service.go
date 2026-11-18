package auth

import (
	"errors"
	"time"

	"app/internal/database"
	"app/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	GetProfile(userID uint) (*UserResponse, error)
	UpdateProfile(userID uint, req *UpdateProfileRequest) (*UserResponse, error)
}

type service struct {
	repo      Repository
	secretKey string
}

func NewService(repo Repository, secretKey string) Service {
	return &service{repo: repo, secretKey: secretKey}
}

func (s *service) generateToken(user *database.User) string {
	claims := &middleware.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(s.secretKey))
	return tokenStr
}

func (s *service) Register(req *RegisterRequest) (*AuthResponse, error) {
	if existing, _ := s.repo.FindByUsername(req.Username); existing != nil {
		return nil, errors.New("username already in use")
	}

	if existing, _ := s.repo.FindByEmail(req.Email); existing != nil {
		return nil, errors.New("email already in use")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashed),
		Role:     "member",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	token := s.generateToken(user)

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (s *service) Login(req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token := s.generateToken(user)

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (s *service) GetProfile(userID uint) (*UserResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *service) UpdateProfile(userID uint, req *UpdateProfileRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Username != "" && req.Username != user.Username {
		if existing, _ := s.repo.FindByUsername(req.Username); existing != nil {
			return nil, errors.New("username already in use")
		}
		user.Username = req.Username
	}

	if req.Email != "" && req.Email != user.Email {
		if existing, _ := s.repo.FindByEmail(req.Email); existing != nil {
			return nil, errors.New("email already in use")
		}
		user.Email = req.Email
	}

	if req.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user.Password = string(hashed)
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
