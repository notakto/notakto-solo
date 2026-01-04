package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreatePlayer(ctx context.Context, q *db.Queries, uid string, name string, email string, profilePic string) (
	err error,
) {
	start := time.Now()
	err = q.CreatePlayer(ctx, db.CreatePlayerParams{
		Uid:   uid,
		Name:  name,
		Email: email,
		ProfilePic: pgtype.Text{
			String: profilePic,
			Valid:  profilePic != "",
		},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreatePlayer took %v, err: %v", time.Since(start), err)
	}
	return err
}
