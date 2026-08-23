package usecase

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/logic"
	"github.com/rakshitg600/notakto-solo/store"
)

func EnsureUpdateUsername(ctx context.Context, pool *pgxpool.Pool, username string) (string, int, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", http.StatusUnauthorized, errors.New("missing or invalid uid in context")
	}
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return "", http.StatusBadRequest, errors.New("username is required")
	}

	if err := logic.ValidateUsername(trimmed); err != nil {
		return "", http.StatusBadRequest, err
	}

	queries := db.New(pool)
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("EnsureUpdateUsername: Failed to rollback transaction: %v", err)
		}
	}(tx, ctx)

	qtx := queries.WithTx(tx)

	exists, err := store.CheckUsernameExists(ctx, qtx, trimmed)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	if exists {
		return "", http.StatusConflict, errors.New("username already exists")
	}

	player, err := store.UpdatePlayerUsername(ctx, qtx, trimmed)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", http.StatusConflict, errors.New("username already exists")
		}
		return "", http.StatusInternalServerError, err
	}

	if err := tx.Commit(ctx); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", http.StatusConflict, errors.New("username already exists")
		}
		return "", http.StatusInternalServerError, err
	}

	return player.Username, http.StatusOK, nil
}
