package config

import "slices"

type CoinPackage struct {
	ID          string
	Coins       int32
	AmountCents int32
	Currency    string
}

var coinPackages = []CoinPackage{
	{
		ID:          "pkg_500",
		Coins:       500,
		AmountCents: 99,
		Currency:    "USD",
	},
	{
		ID:          "pkg_1200",
		Coins:       1200,
		AmountCents: 199,
		Currency:    "USD",
	},
	{
		ID:          "pkg_3000",
		Coins:       3000,
		AmountCents: 499,
		Currency:    "USD",
	},
}

var coinPackagesByID = func() map[string]CoinPackage {
	packages := make(map[string]CoinPackage, len(coinPackages))
	for _, pkg := range coinPackages {
		packages[pkg.ID] = pkg
	}
	return packages
}()

func ListCoinPackages() []CoinPackage {
	return slices.Clone(coinPackages)
}

func CoinPackageByID(packageID string) (CoinPackage, bool) {
	pkg, ok := coinPackagesByID[packageID]
	return pkg, ok
}
