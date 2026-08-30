package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/imagekitservice"
	"github.com/rakshitg600/notakto-solo/logic"
	"github.com/rakshitg600/notakto-solo/store"
	"github.com/redis/go-redis/v9"
)

const (
	usernameAllocationLockKeyPrefix = "lock:username-allocation:"
	usernameAllocationLockTTL       = 12 * time.Second
)

type LoginResult struct {
	ProfilePic string
	Name       string
	Email      string
	Username   string
	IsNew      bool
}

func EnsureLogin(
	ctx context.Context,
	pool *pgxpool.Pool,
	authClient *auth.Client,
	rdb *redis.Client,
) (LoginResult, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return LoginResult{}, errors.New("missing or invalid uid in context")
	}
	// STEP 1: Try existing session
	queries := db.New(pool)
	existing, err := store.GetPlayerById(ctx, queries)
	if err == nil && existing.Uid != "" {
		result := LoginResult{
			Name:     existing.Name,
			Email:    existing.Email,
			Username: existing.Username,
		}
		if existing.ProfilePic.Valid {
			result.ProfilePic = existing.ProfilePic.String
		}
		if existing.ProfileImageFileID.Valid && existing.ProfileImageFilePath.Valid {
			profilePic, err := imagekitservice.BuildURL(existing.ProfileImageFilePath.String)
			if err != nil {
				return LoginResult{}, err
			}
			result.ProfilePic = profilePic
		}
		return result, nil
	}
	if err == nil && existing.Uid == "" {
		return LoginResult{}, errors.New("empty player returned from db")
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, err
	}
	// STEP 2: Fetch profile from Firebase
	name, email, profilePic, err := GetFirebaseUserProfile(ctx, authClient)
	if err != nil {
		return LoginResult{}, err
	}
	signUp, err := loadSignUpConfig(ctx, queries)
	if err != nil {
		return LoginResult{}, err
	}
	usernameLists, err := loadUsernameWordLists(ctx, queries)
	if err != nil {
		return LoginResult{}, err
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return LoginResult{}, err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("EnsureLogin: Failed to rollback transaction: %v", err)
		}
	}(tx, ctx)

	qtx := queries.WithTx(tx)
	// STEP 3: Create new player
	username, err := generateAvailableUsername(ctx, qtx, usernameLists)
	if err != nil {
		return LoginResult{}, err
	}

	lockKey := usernameAllocationLockKeyPrefix + username
	unlock, err := logic.AcquireRedisLock(ctx, rdb, lockKey, usernameAllocationLockTTL)
	if errors.Is(err, logic.ErrRedisLockAlreadyHeld) {
		return LoginResult{}, err
	}
	if err != nil {
		return LoginResult{}, err
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := unlock(unlockCtx); err != nil {
			log.Printf("EnsureLogin: Failed to release username allocation lock: %v", err)
		}
	}()

	err = store.CreatePlayer(ctx, qtx, name, email, profilePic, username)
	if err != nil {
		return LoginResult{}, err
	}
	// STEP 4: Create Wallet for player
	err = store.CreateWallet(ctx, qtx, signUp)
	if err != nil {
		return LoginResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LoginResult{}, err
	}
	// STEP 5: Return values
	return LoginResult{
		ProfilePic: profilePic,
		Name:       name,
		Email:      email,
		Username:   username,
		IsNew:      true,
	}, nil
}
