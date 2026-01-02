package usecase

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/rakshitg600/notakto-solo/config"
	db "github.com/rakshitg600/notakto-solo/db/generated"
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
	start := time.Now()
	GetPlayerByIdCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	existing, err := q.GetPlayerById(GetPlayerByIdCtx, uid)
	log.Printf("GetPlayerById took %v, err: %v", time.Since(start), err)
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
	createPlayerCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.CreatePlayer(createPlayerCtx, db.CreatePlayerParams{
		Uid:   uid,
		Name:  name,
		Email: email,
		ProfilePic: sql.NullString{
			String: profile_pic,
			Valid:  profile_pic != "",
		},
	})
	if err != nil {
		return "", "", "", true, err
	}
	// STEP 4: Create Wallet for player
	createWalletCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = q.CreateWallet(createWalletCtx, db.CreateWalletParams{
		Uid: uid,
		Coins: sql.NullInt32{
			Int32: config.Wallet.InitialCoins,
			Valid: true,
		},
		Xp: sql.NullInt32{
			Int32: config.Wallet.InitialXP,
			Valid: true,
		},
	})
	if err != nil {
		return "", "", "", true, err
	}

	// STEP 5: Return values
	return profile_pic, name, email, true, nil
}
