package concurrency

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisGuard struct {
	rdb       *redis.Client
	ttl       time.Duration
	wait      time.Duration // 0 = reject mode
	prefix    string
	pollDelay time.Duration
}

func NewRedisGuard(
	rdb *redis.Client,
	ttl time.Duration,
	wait time.Duration,
	prefix string,
) *RedisGuard {
	if prefix == "" {
		prefix = "lock:uid:"
	}
	return &RedisGuard{
		rdb:       rdb,
		ttl:       ttl,
		wait:      wait,
		prefix:    prefix,
		pollDelay: 50 * time.Millisecond,
	}
}
