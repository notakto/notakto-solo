package usecase

import (
	"context"
	"errors"
	"log"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureLogin(ctx context.Context, pool *pgxpool.Pool, authClient *auth.Client) (profilePic string, name string, email string, username string, isNew bool, err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", "", "", "", false, errors.New("missing or invalid uid in context")
	}
	// STEP 1: Try existing session
	queries := db.New(pool)
	existing, err := store.GetPlayerById(ctx, queries)
	if err == nil && existing.Uid != "" {
		name = existing.Name
		email = existing.Email
		username = existing.Username
		if existing.ProfilePic.Valid {
			profilePic = existing.ProfilePic.String
		} else {
			profilePic = ""
		}
		return profilePic, name, email, username, false, nil
	}
	if err == nil && existing.Uid == "" {
		return "", "", "", "", false, errors.New("empty player returned from db")
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", "", false, err
	}
	// STEP 2: Fetch profile from Firebase
	name, email, profilePic, err = GetFirebaseUserProfile(ctx, authClient)
	if err != nil {
		return "", "", "", "", true, err
	}
	signUp, err := loadSignUpConfig(ctx, queries)
	if err != nil {
		return "", "", "", "", true, err
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return "", "", "", "", true, err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Printf("EnsureLogin: Failed to rollback transaction: %v", err)
		}
	}(tx, ctx)

	qtx := queries.WithTx(tx)
	// STEP 3: Create new player
	username, _, _ = strings.Cut(email, "@")
	err = store.CreatePlayer(ctx, qtx, name, email, profilePic, username)
	if err != nil {
		return "", "", "", "", true, err
	}
	// STEP 4: Create Wallet for player
	err = store.CreateWallet(ctx, qtx, signUp)
	if err != nil {
		return "", "", "", "", true, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", "", "", "", true, err
	}
	// STEP 5: Return values
	return profilePic, name, email, username, true, nil
}
