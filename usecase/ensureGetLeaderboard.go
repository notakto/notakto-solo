package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

type LeaderboardEntry struct {
	Rank     int32
	Username string
	XP       int32
}

func EnsureGetLeaderboard(ctx context.Context, pool *pgxpool.Pool) ([]LeaderboardEntry, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return nil, errors.New("missing or invalid uid in context")
	}

	q := db.New(pool)
	rows, err := store.GetLeaderboard(ctx, q)
	if err != nil {
		return nil, err
	}

	leaderboard := make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		leaderboard = append(leaderboard, LeaderboardEntry{
			Rank:     row.Rank,
			Username: row.Username,
			XP:       row.Xp,
		})
	}
	return leaderboard, nil
}
