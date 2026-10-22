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

type LeaderboardResult struct {
	Leaderboard []LeaderboardEntry
	CurrentUser *LeaderboardEntry
}

func EnsureGetLeaderboard(ctx context.Context, pool *pgxpool.Pool) (*LeaderboardResult, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return nil, errors.New("missing or invalid uid in context")
	}

	q := db.New(pool)
	rows, err := store.GetLeaderboard(ctx, q)
	if err != nil {
		return nil, err
	}
	currentWallet, err := store.GetWalletByPlayerId(ctx, q)
	if err != nil {
		return nil, err
	}
	playersWithMoreXP, err := store.CountWalletsWithMoreXP(ctx, q, currentWallet.Xp)
	if err != nil {
		return nil, err
	}

	result := &LeaderboardResult{
		Leaderboard: make([]LeaderboardEntry, 0, 10),
	}
	players := make(map[string]db.Player, len(rows)+1)
	playerByUID := func(playerUID string) (db.Player, error) {
		if player, ok := players[playerUID]; ok {
			return player, nil
		}
		player, err := q.GetPlayerById(ctx, playerUID)
		if err != nil {
			return db.Player{}, err
		}
		players[playerUID] = player
		return player, nil
	}
	for i, row := range rows {
		player, err := playerByUID(row.Uid)
		if err != nil {
			return nil, err
		}
		rank := int32(1)
		if i > 0 {
			rank = result.Leaderboard[i-1].Rank
			if row.Xp != rows[i-1].Xp {
				rank = int32(i + 1)
			}
		}
		result.Leaderboard = append(result.Leaderboard, LeaderboardEntry{
			Rank:     rank,
			Username: player.Username,
			XP:       row.Xp,
		})
	}
	currentPlayer, err := playerByUID(uid)
	if err != nil {
		return nil, err
	}
	result.CurrentUser = &LeaderboardEntry{
		Rank:     int32(playersWithMoreXP + 1),
		Username: currentPlayer.Username,
		XP:       currentWallet.Xp,
	}
	return result, nil
}
