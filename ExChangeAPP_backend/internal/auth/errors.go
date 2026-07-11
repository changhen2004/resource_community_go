package auth

import "errors"

var (
	ErrUsernameExists      = errors.New("username already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
)
