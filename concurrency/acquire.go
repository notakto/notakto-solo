package concurrency

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (g *RedisGuard) Acquire(ctx context.Context, key string) (*Result, error) {
	lockKey := g.prefix + key
	token := uuid.NewString()
	deadline := time.Now().Add(g.wait)

	for {
		ok, err := g.rdb.SetNX(
			ctx,
			lockKey,
			token,
			g.ttl,
		).Result()

		if err != nil {
			return nil, err
		}

		if ok {
			return &Result{Acquired: true}, nil
		}

		if g.wait == 0 || time.Now().After(deadline) {
			return &Result{Acquired: false}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(g.pollDelay):
		}
	}
}
