package cachekey

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ArticleListPrefix   = "articles:list:"
	ArticleDetailPrefix = "articles:detail:"
	ArticleHotPrefix    = "articles:hot:"
	ArticleHotZSetKey   = "articles:hot:zset"
	PointsSummaryPrefix = "points:summary:"
)

const (
	ArticleListTTL   = 5 * time.Minute
	ArticleDetailTTL = 10 * time.Minute
	ArticleHotTTL    = 3 * time.Minute
	PointsSummaryTTL = 2 * time.Minute
)

func ArticleListKey(page, pageSize int, sort, keyword, tag string) string {
	return fmt.Sprintf(
		"%spage=%d:size=%d:sort=%s:keyword=%s:tag=%s",
		ArticleListPrefix,
		page,
		pageSize,
		sort,
		keyword,
		tag,
	)
}

func ArticleDetailKey(articleID string) string {
	return ArticleDetailPrefix + articleID
}

func ArticleHotKey(limit int) string {
	return fmt.Sprintf("%slimit=%d", ArticleHotPrefix, limit)
}

func PointsSummaryKey(userID uint) string {
	return fmt.Sprintf("%s%d", PointsSummaryPrefix, userID)
}

func DeleteByPrefix(ctx context.Context, redisDB *redis.Client, prefix string) {
	if redisDB == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	iter := redisDB.Scan(ctx, 0, prefix+"*", 0).Iterator()
	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		_ = redisDB.Del(ctx, keys...).Err()
	}
}

func DeleteKeys(ctx context.Context, redisDB *redis.Client, keys ...string) {
	if redisDB == nil || len(keys) == 0 {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	_ = redisDB.Del(ctx, keys...).Err()
}
