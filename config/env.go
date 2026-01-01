package config

import (
	"errors"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type EnvMode string

const (
	EnvPreview EnvMode = "preview"
	EnvProd    EnvMode = "prod"
)

var (
	envStore sync.Map
	envMode  EnvMode
	initOnce sync.Once
)

// InitEnv must be called once at startup
func InitEnv() error {
	var initErr error
	initOnce.Do(func() {
		// Load .env for local/dev
		_ = godotenv.Load()

		// Detect environment mode
		if os.Getenv("RENDER_GIT_PULL_REQUEST") != "" {
			envMode = EnvPreview
		} else {
			envMode = EnvProd
		}

		// Common variables
		if err := load("PORT", "1323"); err != nil {
			initErr = err
			return
		}

		// Logical variables (callers never care about dev/prod)
		if err := loadResolved("DATABASE_URL"); err != nil {
			initErr = err
			return
		}

		if err := loadResolved("FIREBASE_API_KEY"); err != nil {
			initErr = err
			return
		}

	})
	return initErr
}

// Resolve logical key -> actual env key
func resolveKey(key string) string {
	if envMode == EnvPreview {
		switch key {
		case "DATABASE_URL":
			return "DATABASE_DEV_URL"
		case "FIREBASE_API_KEY":
			return "FIREBASE_DEV_API_KEY"
		}
	}
	return key
}

func loadResolved(key string) error {
	return load(resolveKey(key))
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

// GetEnv returns the resolved env value
func GetEnv(key string) (string, bool) {
	actualKey := resolveKey(key)
	val, ok := envStore.Load(actualKey)
	if !ok {
		return "", false
	}
	return val.(string), true
}

// MustGetEnv panics if env var is missing (recommended for required config)
func MustGetEnv(key string) string {
	val, ok := GetEnv(key)
	if !ok {
		panic("missing environment variable: " + key)
	}
	return val
}
