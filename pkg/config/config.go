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

	// DefaultCertPath is the default path to certificates
	DefaultCertPath = "./etc"

	// DefaultDBPath is the default path to the database
	DefaultDBPath = "./etc/waffy.db"

	// DefaultRPCName is the hostname of the RPC
	DefaultRPCName = "waffy.local"
)

// Version is the version of the software
var Version = "0.0.1"

// Config is the main configuration for waffy
type Config struct {
	// API Listener is the address the API should listen on
	APIListen string

	// CertPath is the path to certificates for the system
	CertPath string

	// DBPath is the path to the internal database
	DBPath string

	// RPCName is the hostname of the RPC client
	RPCName string

	// Version is the version currently running
	Version string
}

var cfg *Config

func init() {
	c, err := godotenv.Read()
	if err != nil {
		panic("cannot read configuration environment")
	}

	cfg = &Config{
		APIListen: getEnv("WAFFY_API_LISTEN", c, DefaultAPIListen),
		CertPath:  getEnv("WAFFY_CERT_PATH", c, DefaultCertPath),
		DBPath:    getEnv("WAFFY_DB_PATH", c, DefaultDBPath),
		RPCName:   getEnv("WAFFY_RPC_NAME", c, DefaultRPCName),

		Version: Version,
	}

}

func getEnv(cfg string, c map[string]string, defValue string) string {
	val := os.Getenv(cfg)
	if val == "" {
		var ok bool
		if val, ok = c[cfg]; !ok {
			val = defValue
		}
	}

	return val
}

// Load returns the loaded configuration
func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	return nil, fmt.Errorf("Error reading configuration")
}
