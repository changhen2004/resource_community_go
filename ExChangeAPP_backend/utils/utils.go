package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtSecret        = "secret"
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

const (
	accessTokenTTL  = 24 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

type AuthClaims struct {
	UserID       uint
	Username     string
	TokenType    string
	TokenVersion uint64
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func GenerateJWT(userID uint, username string) (string, error) {
	return GenerateAccessToken(userID, username, 0)
}

func GenerateAccessToken(userID uint, username string, tokenVersion uint64) (string, error) {
	return generateJWT(userID, username, tokenVersion, accessTokenType, accessTokenTTL)
}

func GenerateRefreshToken(userID uint, username string, tokenVersion uint64) (string, error) {
	return generateJWT(userID, username, tokenVersion, refreshTokenType, refreshTokenTTL)
}

func generateJWT(userID uint, username string, tokenVersion uint64, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":       userID,
		"username":      username,
		"token_type":    tokenType,
		"token_version": tokenVersion,
		"exp":           now.Add(ttl).Unix(),
		"iat":           now.Unix(),
		"jti":           fmt.Sprintf("%d", now.UnixNano()),
	})

	return token.SignedString([]byte(jwtSecret))
}

func CheckPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ParseJWT(tokenString string) (AuthClaims, error) {
	return parseJWT(tokenString, "")
}

func ParseAccessToken(tokenString string) (AuthClaims, error) {
	return parseJWT(tokenString, accessTokenType)
}

func ParseRefreshToken(tokenString string) (AuthClaims, error) {
	return parseJWT(tokenString, refreshTokenType)
}

func parseJWT(tokenString, expectedTokenType string) (AuthClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return AuthClaims{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return AuthClaims{}, errors.New("user_id claim is not a number")
		}
		username, ok := claims["username"].(string)
		if !ok {
			return AuthClaims{}, errors.New("username claim is not a string")
		}
		tokenType, ok := claims["token_type"].(string)
		if !ok || tokenType == "" {
			return AuthClaims{}, errors.New("token_type claim is not a string")
		}
		if expectedTokenType != "" && tokenType != expectedTokenType {
			return AuthClaims{}, errors.New("token type is not valid")
		}
		tokenVersionFloat, ok := claims["token_version"].(float64)
		if !ok {
			return AuthClaims{}, errors.New("token_version claim is not a number")
		}
		return AuthClaims{
			UserID:       uint(userIDFloat),
			Username:     username,
			TokenType:    tokenType,
			TokenVersion: uint64(tokenVersionFloat),
		}, nil
	}
	return AuthClaims{}, errors.New("token is not valid")
}
