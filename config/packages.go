package config

const CoinPackagesKey = "coin_packages"

type CoinPackage struct {
	ID          string `json:"id"`
	PackageName string `json:"package_name"`
	Coins       int32  `json:"coins"`
	VisualCoins int32  `json:"visual_coins"`
	AmountCents int32  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

var defaultCoinPackages = []CoinPackage{
	{
		ID:          "pkg_500",
		PackageName: "Starter Pack",
		Coins:       500,
		VisualCoins: 2,
		AmountCents: 99,
		Currency:    "USD",
	},
	{
		ID:          "pkg_1200",
		PackageName: "Tactical Pack",
		Coins:       1200,
		VisualCoins: 3,
		AmountCents: 199,
		Currency:    "USD",
	},
	{
		ID:          "pkg_3000",
		PackageName: "Champion Pack",
		Coins:       3000,
		VisualCoins: 4,
		AmountCents: 499,
		Currency:    "USD",
	},
}

func DefaultCoinPackages() []CoinPackage {
	return defaultCoinPackages
}

func CoinPackageByID(packages []CoinPackage, packageID string) (CoinPackage, bool) {
	for _, pkg := range packages {
		if pkg.ID == packageID {
			return pkg, true
		}
	}
	return CoinPackage{}, false
}
