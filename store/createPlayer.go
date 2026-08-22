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

func CreatePlayer(ctx context.Context, q *db.Queries, name string, email string, profilePic string, username string) (err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return errors.New("missing or invalid uid in context")
	}
	start := time.Now()
	err = q.CreatePlayer(ctx, db.CreatePlayerParams{
		Uid:   uid,
		Name:  name,
		Email: email,
		ProfilePic: pgtype.Text{
			String: profilePic,
			Valid:  profilePic != "",
		},
		Username: username,
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreatePlayer took %v, err: %v", time.Since(start), err)
	}
	return err
}
