package valkey

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Client = redis.Client

func NewClient(addr, password string, db int) *Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,            // address of Valkey server, e.g., "localhost:6379"
		Password:     password,        // password, if any ( empty string if none)
		DB:           db,              // same valkey instance can have multiple DBs from 0 to 15, default is 0 (not to switch)
		DialTimeout:  5 * time.Second, // timeout for establishing new connections, eg., if connection not established in 5 seconds then error out
		ReadTimeout:  3 * time.Second, // timeout for reading data
		WriteTimeout: 3 * time.Second, // timeout for writing data
	})
}
