package usecase

import (
	"context"

	"github.com/rakshitg600/notakto-solo/config"
)

func EnsureGetAllPackages(_ context.Context) []config.CoinPackage {
	return config.ListCoinPackages()
}
