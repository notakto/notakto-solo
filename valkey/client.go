package valkey

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Client = redis.Client

func NewClient(addr, password string, db int) *Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}
