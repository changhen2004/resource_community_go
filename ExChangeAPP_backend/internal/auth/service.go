package auth

import (
	"errors"
	"exchangeapp/utils"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(req RegisterRequest) (AuthResponse, error) {
	_, err := s.repo.FindByUsername(req.Username)
	if err == nil {
		return AuthResponse{}, ErrUsernameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthResponse{}, err
	}

	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	user := &User{
		Username: req.Username,
		Password: hashedPwd,
	}
	if err := s.repo.Create(user); err != nil {
		return AuthResponse{}, err
	}

	token, err := utils.GenerateJWT(user.ID, req.Username)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: token}, nil
}

func (s *Service) Login(req LoginRequest) (AuthResponse, error) {
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AuthResponse{}, ErrUserNotFound
		}
		return AuthResponse{}, err
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return AuthResponse{}, ErrInvalidPassword
	}

	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: token}, nil
}
