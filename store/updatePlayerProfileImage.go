package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func UpdatePlayerProfileImage(ctx context.Context, q *db.Queries, fileID string, filePath string) (player db.Player, err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return db.Player{}, errors.New("missing or invalid uid in context")
	}

	start := time.Now()
	player, err = q.UpdatePlayerProfileImage(ctx, db.UpdatePlayerProfileImageParams{
		Uid: uid,
		ProfileImageFileID: pgtype.Text{
			String: fileID,
			Valid:  true,
		},
		ProfileImageFilePath: pgtype.Text{
			String: filePath,
			Valid:  true,
		},
	})
	if time.Since(start) > 2*time.Second {
		log.Printf("UpdatePlayerProfileImage took %v, err: %v", time.Since(start), err)
	}
	return player, err
}
