package config

import (
	"errors"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	envStore sync.Map
	initOnce sync.Once
	initErr  error
)

// InitEnv loads env; call once at startup.
func InitEnv() error {
	initOnce.Do(func() {
		// Load .env for local use and let it override stale shell exports.
		// This keeps local config authoritative when both exist.
		_ = godotenv.Overload()

		// Common variables
		if err := load("PORT", "1323"); err != nil {
			initErr = err
			return
		}

		if err := load("DATABASE_URL"); err != nil {
			initErr = err
			return
		}

		if err := load("FIREBASE_CREDENTIALS_JSON"); err != nil {
			initErr = err
			return
		}

		if err := load("VALKEY_URL"); err != nil {
			initErr = err
			return
		}

		if err := load("NOWPAYMENTS_API_KEY"); err != nil {
			initErr = err
			return
		}

		if err := load("NOWPAYMENTS_IPN_SECRET"); err != nil {
			initErr = err
			return
		}

		if err := load("KEEPALIVE_TOKEN"); err != nil {
			initErr = err
			return
		}

		if err := load("IMAGEKIT_PUBLIC_KEY"); err != nil {
			initErr = err
			return
		}

		if err := load("IMAGEKIT_PRIVATE_KEY"); err != nil {
			initErr = err
			return
		}

		if err := load("IMAGEKIT_URL_ENDPOINT"); err != nil {
			initErr = err
			return
		}

	})
	return initErr
}

func load(key string, defaults ...string) error {
	val := os.Getenv(key)
	if val == "" {
		if len(defaults) > 0 {
			val = defaults[0]
		} else {
			return errors.New("missing required environment variable: " + key)
		}
	}
	envStore.Store(key, val)
	return nil
}

// GetEnv returns the loaded env value.
func GetEnv(key string) (string, bool) {
	val, ok := envStore.Load(key)
	if !ok {
		return "", false
	}
	return val.(string), true
}

// MustGetEnv returns the env value or panics if missing.
func MustGetEnv(key string) string {
	val, ok := GetEnv(key)
	if !ok {
		panic("missing environment variable: " + key)
	}
	return val
}
