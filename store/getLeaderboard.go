package store

import (
	"context"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func GetLeaderboard(ctx context.Context, q *db.Queries) ([]db.GetLeaderboardRow, error) {
	start := time.Now()
	leaderboard, err := q.GetLeaderboard(ctx)
	if time.Since(start) > 2*time.Second {
		log.Printf("GetLeaderboard took %v, err: %v", time.Since(start), err)
	}
	if err != nil {
		return nil, err
	}
	return leaderboard, nil
}
