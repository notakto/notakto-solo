package usecase

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

// EnsureLogin authenticates a user and ensures a corresponding player record and wallet exist.
//
// If a player with the provided uid exists, it returns that player's profile picture URL (empty if none),
// name, and email with the new-player flag set to `false`. If no player exists, it verifies the provided
// ID token, creates a new player record and an associated wallet initialized from configuration, and returns
// the profile picture URL, name, email with the new-player flag set to `true`. Database and verification
// errors are propagated; an empty player row from the database is returned as an explicit error.
//
// It returns the profile picture URL, name, email, `true` if a new player was created, `false` otherwise, and any error.
func EnsureLogin(ctx context.Context, q *db.Queries, uid string, idToken string) (profile_pic string, name string, email string, isNew bool, err error) {
	// STEP 1: Try existing session
	existing, err := store.GetPlayerById(ctx, q, uid)
	if err == nil && existing.Uid != "" {
		name = existing.Name
		email = existing.Email
		if existing.ProfilePic.Valid {
			profile_pic = existing.ProfilePic.String
		} else {
			profile_pic = ""
		}
		return profile_pic, name, email, false, nil
	}
	if err == nil && existing.Uid == "" {
		return "", "", "", false, errors.New("empty player returned from db")
	}
	if err != nil && err != sql.ErrNoRows {
		return "", "", "", false, err
	}
	// STEP 2: Fetch from Firebase
	uid, name, email, profile_pic, err = VerifyFirebaseToken(ctx, idToken)
	if err != nil {
		return "", "", "", true, err
	}

	// STEP 3: Create new player
	err = store.CreatePlayer(ctx, q, uid, name, email, profile_pic)
	if err != nil {
		return "", "", "", true, err
	}
	// STEP 4: Create Wallet for player

	err = store.CreateWallet(ctx, q, uid, name, email, profile_pic)
	if err != nil {
		return "", "", "", true, err
	}
	// STEP 5: Return values
	return profile_pic, name, email, true, nil
}
