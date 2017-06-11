// Package config contains configuration for waffy
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	// DefaultAPIListen is the default listen address for the RPC
	DefaultAPIListen = "0.0.0.0:8500"

	// DefaultCertPath is the default path to certificates
	DefaultCertPath = "./etc"

	// DefaultDBPath is the default path to the database
	DefaultDBPath = "./etc/waffy.db"

	// DefaultRPCName is the hostname of the RPC
	DefaultRPCName = "waffy.local"

	// DefaultRaftDIR is the default Raft configuration directory
	DefaultRaftDIR = "./etc/raft"

	// DefaultRaftListen is the default Raft listen address
	DefaultRaftListen = "127.0.0.1:8501"
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

	// RaftDIR is the directory for Raft peer and log storage
	RaftDIR string

	// RaftListen is the listen address of the Raft consensus
	RaftListen string
}

var cfg *Config

func init() {
	c, err := godotenv.Read()
	if err != nil {
		panic("cannot read configuration environment")
	}

	cfg = &Config{
		APIListen:  getEnv("WAFFY_API_LISTEN", c, DefaultAPIListen),
		CertPath:   getEnv("WAFFY_CERT_PATH", c, DefaultCertPath),
		DBPath:     getEnv("WAFFY_DB_PATH", c, DefaultDBPath),
		RPCName:    getEnv("WAFFY_RPC_NAME", c, DefaultRPCName),
		RaftDIR:    getEnv("WAFFY_RAFT_DIR", c, DefaultRaftDIR),
		RaftListen: getEnv("WAFFY_RAFT_LISTEN", c, DefaultRaftListen),

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

// ensureFile ensures that a file at a given directory exists
func ensureFile(base, filename string, create bool, truncate bool) (*os.File, error) {
	var path string

	dir, fName := filepath.Split(filename)
	if fName != "" {
		dir := filepath.Join(base, dir)
		if err := os.MkdirAll(dir, 0700); err != nil {
			if !os.IsExist(err) || !os.IsNotExist(err) {
				return nil, err
			}
		}

		path = filepath.Join(dir, fName)
	}
	if fName == "" {
		path = filepath.Join(base, filename)

	}

	fileFlags := os.O_RDWR
	if create {
		fileFlags = fileFlags | os.O_CREATE
	}

	if truncate {
		fileFlags = fileFlags | os.O_TRUNC
	}

	f, err := os.OpenFile(path, fileFlags, 0600)
	if err != nil {
		return nil, fmt.Errorf("unable to open path %s: %s", path, err)
	}

	return f, nil
}
