package usecase

import "github.com/rakshitg600/notakto-solo/config"

func EnsureGetAllPackages() []config.CoinPackage {
	return config.ListCoinPackages()
}
