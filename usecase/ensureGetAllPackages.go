package usecase

import "github.com/rakshitg600/notakto-solo/config"

import (
	"context"

	"github.com/rakshitg600/notakto-solo/config"
)

func EnsureGetAllPackages(ctx context.Context) []config.CoinPackage {
	return config.ListCoinPackages()
}
