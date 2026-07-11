package auth

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repo struct {
	db      *gorm.DB
	redisDB *redis.Client
}

func NewRepo(db *gorm.DB, redisDB *redis.Client) *Repo {
	return &Repo{db: db, redisDB: redisDB}
}

func (r *Repo) FindByUsername(username string) (*User, error) {
	var user User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repo) FindByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repo) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *Repo) GetTokenVersion(ctx context.Context, userID uint) (uint64, error) {
	if r.redisDB == nil {
		return 0, nil
	}

	value, err := r.redisDB.Get(ctx, tokenVersionKey(userID)).Uint64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (r *Repo) IncrementTokenVersion(ctx context.Context, userID uint) (uint64, error) {
	if r.redisDB == nil {
		return 1, nil
	}

	return r.redisDB.Incr(ctx, tokenVersionKey(userID)).Uint64()
}

func tokenVersionKey(userID uint) string {
	return fmt.Sprintf("auth:token_version:%d", userID)
}
