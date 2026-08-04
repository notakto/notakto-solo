package config

const SignUpKey = "sign_up"

type SignUpConfig struct {
	InitialCoins int32 `json:"initial_coins"`
	InitialXP    int32 `json:"initial_xp"`
}

var defaultSignUpConfig = SignUpConfig{
	InitialCoins: 1000,
	InitialXP:    0,
}

func DefaultSignUpConfig() SignUpConfig {
	return defaultSignUpConfig
}
