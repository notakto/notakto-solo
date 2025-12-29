package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucket struct {
	rdb        *redis.Client
	capacity   int64
	refillRate int64 // tokens per second
	prefix     string
}

func NewTokenBucket(
	rdb *redis.Client,
	capacity int64,
	refillRate int64,
	prefix string,
) *TokenBucket {
	if prefix == "" {
		prefix = "ratelimit:token:"
	}
	return &TokenBucket{rdb, capacity, refillRate, prefix}
}

func (t *TokenBucket) Allow(ctx context.Context, key string) (*Result, error) {
	now := time.Now().Unix()
	redisKey := t.prefix + key

	lua := `
local tokens = tonumber(redis.call("GET", KEYS[1]) or ARGV[1])
local last = tonumber(redis.call("GET", KEYS[2]) or ARGV[2])

local delta = math.max(0, ARGV[2] - last)
local refill = delta * ARGV[3]
tokens = math.min(ARGV[1], tokens + refill)

if tokens < 1 then
	return {0, tokens}
end

tokens = tokens - 1
redis.call("SET", KEYS[1], tokens)
redis.call("SET", KEYS[2], ARGV[2])
redis.call("EXPIRE", KEYS[1], 3600)
redis.call("EXPIRE", KEYS[2], 3600)

return {1, tokens}
`

	res, err := t.rdb.Eval(
		ctx,
		lua,
		[]string{redisKey + ":tokens", redisKey + ":ts"},
		t.capacity,
		now,
		t.refillRate,
	).Result()

	if err != nil {
		return nil, err
	}

	values := res.([]interface{})
	allowed := values[0].(int64) == 1
	remaining := int64(values[1].(int64))

	return &Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetIn:   1,
	}, nil
}
