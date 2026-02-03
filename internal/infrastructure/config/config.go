package config

import "os"

// Config holds environment-driven configuration.
type Config struct {
	Addr string
}

// Load reads configuration from environment variables.
func Load() Config {
	addr := os.Getenv("PET_SHOP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{Addr: addr}
}
