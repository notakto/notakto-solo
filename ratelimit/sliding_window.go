package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindow struct {
	rdb    *redis.Client
	limit  int64
	window time.Duration
	prefix string
}

func NewSlidingWindow(
	rdb *redis.Client,
	limit int64,
	window time.Duration,
	prefix string,
) *SlidingWindow {
	if prefix == "" {
		prefix = "ratelimit:sliding:"
	}
	return &SlidingWindow{rdb, limit, window, prefix}
}

func (s *SlidingWindow) Allow(ctx context.Context, key string) (*Result, error) {
	now := time.Now().UnixMilli()
	windowStart := now - s.window.Milliseconds()
	redisKey := s.prefix + key

	pipe := s.rdb.TxPipeline()

	pipe.ZRemRangeByScore(
		ctx,
		redisKey,
		"0",
		strconv.FormatInt(windowStart, 10),
	)

	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now),
		Member: now,
	})
	countCmd := pipe.ZCard(ctx, redisKey)
	pipe.PExpire(ctx, redisKey, s.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	count := countCmd.Val()

	if count > s.limit {
		return &Result{
			Allowed:   false,
			Remaining: 0,
			ResetIn:   int64(s.window.Seconds()),
		}, nil
	}

	return &Result{
		Allowed:   true,
		Remaining: s.limit - count,
		ResetIn:   int64(s.window.Seconds()),
	}, nil
}
