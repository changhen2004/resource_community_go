package article

import (
	"context"
	"strconv"
	"strings"

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

func (r *Repo) Create(article *Article) error {
	return r.db.Create(article).Error
}

func (r *Repo) List(query ListArticlesQuery) ([]Article, error) {
	var articles []Article

	db := r.db.Model(&Article{})
	if query.Keyword != "" {
		db = db.Where("title LIKE ?", "%"+query.Keyword+"%")
	}
	if query.Tag != "" {
		db = r.applyTagFilter(db, query.Tag)
	}

	switch query.Sort {
	case "hot":
		db = db.Order("like_count DESC").Order("view_count DESC").Order("created_at DESC")
	default:
		db = db.Order("created_at DESC")
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

func (r *Repo) FindByID(id string) (*Article, error) {
	var article Article
	if err := r.db.Where("id = ?", id).First(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *Repo) DeleteArticlesCacheByPrefix(ctx context.Context, prefix string) {
	if r.redisDB == nil {
		return
	}

	iter := r.redisDB.Scan(ctx, 0, prefix+"*", 0).Iterator()
	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		_ = r.redisDB.Del(ctx, keys...).Err()
	}
}

func (r *Repo) GetArticlesCache(ctx context.Context, key string) (string, error) {
	if r.redisDB == nil {
		return "", redis.Nil
	}
	return r.redisDB.Get(ctx, key).Result()
}

func (r *Repo) SetArticlesCache(ctx context.Context, key, value string) {
	if r.redisDB == nil {
		return
	}
	_ = r.redisDB.Set(ctx, key, value, 0).Err()
}

func (r *Repo) IncrementLike(ctx context.Context, articleID string) (int, error) {
	if err := r.db.Model(&Article{}).
		Where("id = ?", articleID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).
		Error; err != nil {
		return 0, err
	}

	article, err := r.FindByID(articleID)
	if err != nil {
		return 0, err
	}

	likes := int(article.LikeCount)
	if r.redisDB != nil {
		likeKey := "article:" + articleID + ":like"
		if err := r.redisDB.Set(ctx, likeKey, likes, 0).Err(); err != nil {
			return 0, err
		}
	}
	return likes, nil
}

func (r *Repo) GetLikeCount(ctx context.Context, articleID string) (int, error) {
	if r.redisDB != nil {
		likeKey := "article:" + articleID + ":like"
		value, err := r.redisDB.Get(ctx, likeKey).Result()
		if err == nil {
			return strconv.Atoi(value)
		}
		if err != redis.Nil {
			return 0, err
		}
	}

	article, err := r.FindByID(articleID)
	if err != nil {
		return 0, err
	}
	return int(article.LikeCount), nil
}

func (r *Repo) applyTagFilter(db *gorm.DB, tag string) *gorm.DB {
	normalizedTag := strings.ToLower(strings.TrimSpace(tag))
	if normalizedTag == "" {
		return db
	}

	switch r.db.Dialector.Name() {
	case "mysql":
		return db.Where("FIND_IN_SET(?, LOWER(tags)) > 0", normalizedTag)
	default:
		return db.Where("instr(',' || lower(tags) || ',', ',' || ? || ',') > 0", normalizedTag)
	}
}
