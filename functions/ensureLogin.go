package functions

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func EnsureLogin(ctx context.Context, q *db.Queries, uid string, idToken string) (profilePic, name, email string, isNew bool, err error) {
	// Try existing player with timeout
	start := time.Now()
	fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	existing, err := q.GetPlayerById(fetchCtx, uid)
	log.Printf("GetPlayerById took %v, err: %v", time.Since(start), err)

	switch {
	case err == nil && existing.Uid != "":
		// Player exists
		name = existing.Name
		email = existing.Email
		if existing.ProfilePic.Valid {
			profilePic = existing.ProfilePic.String
		}
		return profilePic, name, email, false, nil

	case errors.Is(err, sql.ErrNoRows):
		// Player does not exist → create new
		log.Printf("Player with uid %s does not exist, creating new player", uid)

		uid, name, email, profilePic, err = VerifyFirebaseToken(idToken)
		if err != nil {
			return "", "", "", true, err
		}

		// Create player
		createPlayerCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err = q.CreatePlayer(createPlayerCtx, db.CreatePlayerParams{
			Uid:   uid,
			Name:  name,
			Email: email,
			ProfilePic: sql.NullString{
				String: profilePic,
				Valid:  profilePic != "",
			},
		}); err != nil {
			return "", "", "", true, err
		}

		// Create wallet
		createWalletCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err = q.CreateWallet(createWalletCtx, db.CreateWalletParams{
			Uid:   uid,
			Coins: sql.NullInt32{Int32: 1000, Valid: true},
			Xp:    sql.NullInt32{Int32: 0, Valid: true},
		}); err != nil {
			return "", "", "", true, err
		}

		return profilePic, name, email, true, nil

	case err != nil:
		// Any other DB error
		return "", "", "", true, err

	default:
		// Fallback (shouldn’t normally hit)
		return "", "", "", true, nil
	}
}
