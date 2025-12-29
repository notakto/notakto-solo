package ratelimit

import "context"

type Result struct {
	Allowed   bool
	Remaining int64
	ResetIn   int64 // seconds
}

type Limiter interface {
	Allow(ctx context.Context, key string) (*Result, error)
}
