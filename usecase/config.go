package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rakshitg600/notakto-solo/config"
	"github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/store"
)

func loadCoinPackages(ctx context.Context, q *db.Queries) ([]config.CoinPackage, error) {
	value, err := store.GetConfigValueByKey(ctx, q, config.CoinPackagesKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return config.DefaultCoinPackages(), nil
		}
		return nil, fmt.Errorf("get %q config: %w", config.CoinPackagesKey, err)
	}
	packages := make([]config.CoinPackage, 0)
	if err := json.Unmarshal(value, &packages); err != nil {
		return nil, fmt.Errorf("decode %s config: %w", config.CoinPackagesKey, err)
	}
	return packages, nil
}

func loadSignUpConfig(ctx context.Context, q *db.Queries) (config.SignUpConfig, error) {
	value, err := store.GetConfigValueByKey(ctx, q, config.SignUpKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return config.DefaultSignUpConfig(), nil
		}
		return config.SignUpConfig{}, fmt.Errorf("get %q config: %w", config.SignUpKey, err)
	}

	signUp := config.SignUpConfig{}
	if err := json.Unmarshal(value, &signUp); err != nil {
		return config.SignUpConfig{}, fmt.Errorf("decode %s config: %w", config.SignUpKey, err)
	}
	return signUp, nil
}
