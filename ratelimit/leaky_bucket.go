package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type LeakyBucket struct {
	rdb      *redis.Client
	rate     int64 // requests per second
	capacity int64
	prefix   string
}

func NewLeakyBucket(
	rdb *redis.Client,
	rate int64,
	capacity int64,
	prefix string,
) *LeakyBucket {
	if prefix == "" {
		prefix = "ratelimit:leaky:"
	}
	return &LeakyBucket{rdb, rate, capacity, prefix}
}

func (l *LeakyBucket) Allow(ctx context.Context, key string) (*Result, error) {
	now := time.Now().Unix()
	redisKey := l.prefix + key

	lua := `
local level = tonumber(redis.call("GET", KEYS[1]) or 0)
local last = tonumber(redis.call("GET", KEYS[2]) or ARGV[1])

local leaked = (ARGV[1] - last) * ARGV[2]
level = math.max(0, level - leaked)

if level >= ARGV[3] then
	return {0, level}
end

level = level + 1
redis.call("SET", KEYS[1], level)
redis.call("SET", KEYS[2], ARGV[1])
redis.call("EXPIRE", KEYS[1], 3600)
redis.call("EXPIRE", KEYS[2], 3600)

return {1, level}
`

	res, err := l.rdb.Eval(
		ctx,
		lua,
		[]string{redisKey + ":level", redisKey + ":ts"},
		now,
		l.rate,
		l.capacity,
	).Result()

	if err != nil {
		return nil, err
	}

	values := res.([]interface{})
	allowed := values[0].(int64) == 1

	return &Result{
		Allowed: allowed,
	}, nil
}
