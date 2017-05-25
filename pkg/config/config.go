// package config contains configuration for waffy
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	// DefaultWAFListen is the default
	DefaultAPIListen = "0.0.0.0:8500"
)

// Config is the main configuration for waffy
type Config struct {
	// API Listener is the address the API should listen on
	APIListen string
}

var cfg *Config

func init() {
	c, err := godotenv.Read()
	if err != nil {
		panic("cannot read configuration environment")
	}

	apiListen := os.Getenv("HERMES_API_LISTEN")
	if apiListen == "" {
		apiListen = DefaultAPIListen
	}

	cfg = &Config{
		APIListen: apiListen,
	}
}

// Load returns the loaded configuration
func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	return nil, fmt.Errorf("Error reading configuration")
}
