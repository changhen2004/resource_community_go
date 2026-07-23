package worker

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	idempotencyProcessingTTL = 10 * time.Minute
	idempotencyDoneTTL       = 24 * time.Hour
)

type IdempotencyStore interface {
	Begin(ctx context.Context, jobID string) (bool, error)
	Complete(ctx context.Context, jobID string) error
	Release(ctx context.Context, jobID string) error
}

type NoopIdempotencyStore struct{}

func (NoopIdempotencyStore) Begin(context.Context, string) (bool, error) {
	return true, nil
}

func (NoopIdempotencyStore) Complete(context.Context, string) error {
	return nil
}

func (NoopIdempotencyStore) Release(context.Context, string) error {
	return nil
}

type RedisIdempotencyStore struct {
	client *redis.Client
	prefix string
}

func NewRedisIdempotencyStore(client *redis.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{
		client: client,
		prefix: "asyncjob:processed:",
	}
}

func (s *RedisIdempotencyStore) Begin(ctx context.Context, jobID string) (bool, error) {
	if s == nil || s.client == nil || jobID == "" {
		return true, nil
	}

	ok, err := s.client.SetNX(ctx, s.prefix+jobID, "processing", idempotencyProcessingTTL).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *RedisIdempotencyStore) Complete(ctx context.Context, jobID string) error {
	if s == nil || s.client == nil || jobID == "" {
		return nil
	}

	return s.client.Set(ctx, s.prefix+jobID, "done", idempotencyDoneTTL).Err()
}

func (s *RedisIdempotencyStore) Release(ctx context.Context, jobID string) error {
	if s == nil || s.client == nil || jobID == "" {
		return nil
	}

	return s.client.Del(ctx, s.prefix+jobID).Err()
}
