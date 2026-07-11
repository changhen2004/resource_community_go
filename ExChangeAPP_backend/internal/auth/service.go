package auth

import (
	"context"
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

	return s.generateTokenPair(user.ID, req.Username)
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

	return s.generateTokenPair(user.ID, user.Username)
}

func (s *Service) Refresh(req RefreshTokenRequest) (AuthResponse, error) {
	claims, err := s.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return AuthResponse{}, ErrInvalidRefreshToken
	}

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AuthResponse{}, ErrUserNotFound
		}
		return AuthResponse{}, err
	}

	return s.generateTokenPair(user.ID, user.Username)
}

func (s *Service) generateTokenPair(userID uint, username string) (AuthResponse, error) {
	tokenVersion, err := s.repo.GetTokenVersion(context.Background(), userID)
	if err != nil {
		return AuthResponse{}, err
	}

	accessToken, err := utils.GenerateAccessToken(userID, username, tokenVersion)
	if err != nil {
		return AuthResponse{}, err
	}

	refreshToken, err := utils.GenerateRefreshToken(userID, username, tokenVersion)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Logout(userID uint) error {
	_, err := s.repo.IncrementTokenVersion(context.Background(), userID)
	return err
}

func (s *Service) ValidateAccessToken(token string) (utils.AuthClaims, error) {
	claims, err := utils.ParseAccessToken(token)
	if err != nil {
		return utils.AuthClaims{}, ErrInvalidAccessToken
	}
	return s.validateClaimsVersion(claims)
}

func (s *Service) ValidateRefreshToken(token string) (utils.AuthClaims, error) {
	claims, err := utils.ParseRefreshToken(token)
	if err != nil {
		return utils.AuthClaims{}, ErrInvalidRefreshToken
	}
	return s.validateClaimsVersion(claims)
}

func (s *Service) validateClaimsVersion(claims utils.AuthClaims) (utils.AuthClaims, error) {
	currentVersion, err := s.repo.GetTokenVersion(context.Background(), claims.UserID)
	if err != nil {
		return utils.AuthClaims{}, err
	}
	if claims.TokenVersion != currentVersion {
		switch claims.TokenType {
		case "refresh":
			return utils.AuthClaims{}, ErrInvalidRefreshToken
		default:
			return utils.AuthClaims{}, ErrInvalidAccessToken
		}
	}
	return claims, nil
}
