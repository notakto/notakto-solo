package config

type WalletDefaults struct {
	InitialCoins int32
	InitialXP    int32
}

var Wallet = WalletDefaults{
	InitialCoins: 1000,
	InitialXP:    0,
}

type CoinPackage struct {
	ID          string
	Coins       int32
	AmountCents int32
}

var CoinPackages = map[string]CoinPackage{
	"pkg_500": {
		ID:          "pkg_500",
		Coins:       500,
		AmountCents: 500,
	},
	"pkg_1200": {
		ID:          "pkg_1200",
		Coins:       1200,
		AmountCents: 1000,
	},
	"pkg_3000": {
		ID:          "pkg_3000",
		Coins:       3000,
		AmountCents: 2000,
	},
}
