package concurrency

import "context"

type Result struct {
	Acquired bool
}

type Guard interface {
	Acquire(ctx context.Context, key string) (*Result, error)
	Release(ctx context.Context, key string) error
}
