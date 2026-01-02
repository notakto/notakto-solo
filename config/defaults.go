package config

type WalletDefaults struct {
	InitialCoins int32
	InitialXP    int32
}

var Wallet = WalletDefaults{
	InitialCoins: 1000,
	InitialXP:    0,
}
