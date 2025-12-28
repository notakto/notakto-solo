package types

import "time"

type RateLimiterConfig struct {
	Limit  int           // max requests
	Window time.Duration // window duration
	Prefix string        // redis key prefix
}
