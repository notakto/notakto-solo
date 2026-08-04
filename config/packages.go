package config

const CoinPackagesKey = "coin_packages"

type CoinPackage struct {
	PackageID      string `json:"packageId"`
	PackageName    string `json:"packageName"`
	Coins          int32  `json:"coins"`
	VisualCoins    int32  `json:"visualCoins"`
	AmountCents    int32  `json:"amountCents"`
	Currency       string `json:"currency"`
	DefaultPackage bool   `json:"defaultPackage"`
}

var defaultCoinPackages = []CoinPackage{
	{
		PackageID:      "pkg_500",
		PackageName:    "Starter Pack",
		Coins:          500,
		VisualCoins:    2,
		AmountCents:    99,
		Currency:       "USD",
		DefaultPackage: false,
	},
	{
		PackageID:      "pkg_1200",
		PackageName:    "Tactical Pack",
		Coins:          1200,
		VisualCoins:    3,
		AmountCents:    199,
		Currency:       "USD",
		DefaultPackage: true,
	},
	{
		PackageID:      "pkg_3000",
		PackageName:    "Champion Pack",
		Coins:          3000,
		VisualCoins:    4,
		AmountCents:    499,
		Currency:       "USD",
		DefaultPackage: false,
	},
}

func DefaultCoinPackages() []CoinPackage {
	return defaultCoinPackages
}

func CoinPackageByID(packages []CoinPackage, packageID string) (CoinPackage, bool) {
	for _, pkg := range packages {
		if pkg.PackageID == packageID {
			return pkg, true
		}
	}
	return CoinPackage{}, false
}
