package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dobyte/due/v2/utils/xconv"
	"github.com/redis/go-redis/v9"
)

type store struct {
	redis  redis.UniversalClient
	prefix string
}

func (s *store) Get(ctx context.Context, key interface{}) (interface{}, error) {
	value, err := s.redis.Get(ctx, fmt.Sprintf("%s:%s", s.prefix, xconv.String(key))).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return value, err
}

func (s *store) Set(ctx context.Context, key interface{}, value interface{}, duration time.Duration) error {
	return s.redis.Set(ctx, fmt.Sprintf("%s:%s", s.prefix, xconv.String(key)), value, duration).Err()
}

func (s *store) Remove(ctx context.Context, keys ...interface{}) (value interface{}, err error) {
	list := make([]string, 0, len(keys))
	for _, key := range keys {
		list = append(list, fmt.Sprintf("%s:%s", s.prefix, xconv.String(key)))
	}

	return s.redis.Del(ctx, list...).Result()
}
