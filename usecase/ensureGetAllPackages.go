package usecase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/config"
	db "github.com/rakshitg600/notakto-solo/db/generated"
)

func EnsureGetAllPackages(ctx context.Context, pool *pgxpool.Pool) ([]config.CoinPackage, error) {
	return loadCoinPackages(ctx, db.New(pool))
}
