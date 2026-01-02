package store

import (
	"context"
	"database/sql"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func CreatePlayer(ctx context.Context, q *db.Queries, uid string, name string, email string, profile_pic string) (
	err error,
) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.CreatePlayer(ctx, db.CreatePlayerParams{
		Uid:   uid,
		Name:  name,
		Email: email,
		ProfilePic: sql.NullString{
			String: profile_pic,
			Valid:  profile_pic != "",
		},
	})
	if time.Since(start) > 2*time.Second {
		//logging slow DB calls
		log.Printf("CreatePlayer took %v, err: %v", time.Since(start), err)
	}
	return err
}
