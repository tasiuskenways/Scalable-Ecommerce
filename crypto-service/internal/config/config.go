package config

import (
	"os"
)

type Config struct {
	HybridEncryption HybridEncryptionConfig
	AppEnv    string
	AppPort   string
}

type HybridEncryptionConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
}

func Load() *Config {

	return &Config{
		HybridEncryption: HybridEncryptionConfig{
			PrivateKeyPath: getEnv("HYBRID_ENCRYPTION_PRIVATE_KEY_PATH", "app/keys/private.pem"),
			PublicKeyPath:  getEnv("HYBRID_ENCRYPTION_PUBLIC_KEY_PATH", "app/keys/public.pem"),
		},
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}