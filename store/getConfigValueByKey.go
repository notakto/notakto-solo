package store

import (
	"context"
	"log"
	"time"

	"github.com/rakshitg600/notakto-solo/db/generated"
)

func GetConfigValueByKey(ctx context.Context, q *db.Queries, key string) ([]byte, error) {
	start := time.Now()
	value, err := q.GetConfigValueByKey(ctx, key)
	if time.Since(start) > 2*time.Second {
		log.Printf("GetConfigValueByKey took %v, err: %v", time.Since(start), err)
	}
	return value, err
}
