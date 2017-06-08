package config

import (
	"crypto/rsa"
	"crypto/x509"

	"encoding/json"
	"fmt"
	"os"

	"github.com/unerror/waffy/pkg/services/protos/users"
)

const (
	// ClientConfigDir is the directory where client configuration is stored by default
	ClientConfigDir = "$HOME/.waffy"
)

// ClientConfig is the configuration for the client to access RPC
type ClientConfig struct {
	Server  string            `json:"server"`
	User    *users.User       `json:"user"`
	Pubkey  *x509.Certificate `json:"-"`
	Privkey *rsa.PrivateKey   `json:"-"`
}

// CreateClientConfig creates an RPC client configuration stored on disk
func CreateClientConfig(server string, user *users.User, pubkey *x509.Certificate, privkey *rsa.PrivateKey) (*ClientConfig, error) {
	path := os.ExpandEnv(ClientConfigDir)
	if err := os.MkdirAll(path, 0700); err != nil {
		if !os.IsExist(err) {
			return nil, fmt.Errorf("unable to create config directory %s: %s", path, err)
		}
	}

	config := fmt.Sprintf("%s/%s", path, user.Email)
	if err := os.Mkdir(config, 0700); err != nil {
		if !os.IsExist(err) {
			return nil, fmt.Errorf("unable to create user config directory: %s: %s", config, err)
		}
	}

	c := &ClientConfig{
		Server:  server,
		User:    user,
		Pubkey:  pubkey,
		Privkey: privkey,
	}

	clientCfg, err := ensureFile(config, "waffy.json")
	if err != nil {
		return nil, err
	}

	enc := json.NewEncoder(clientCfg)
	if err := enc.Encode(c); err != nil {
		return nil, fmt.Errorf("unable to save client configuration: %s", err)
	}

	err = SaveClientCert(config, user.Email, pubkey, privkey)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// LoadClientConfig loads a client configuration from the filesystem
func LoadClientConfig(path, email string) (*ClientConfig, error) {
	f, err := ensureFile(path, "waffy.json")
	if err != nil {
		return nil, fmt.Errorf("unable to load client configuration: %s", err)
	}

	dec := json.NewDecoder(f)
	c := ClientConfig{}
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("unable to decode client configuration: %s", err)
	}

	return &c, nil
}

// SaveClientCert saves a client Certificate to the filesystem
func SaveClientCert(path, email string, c *x509.Certificate, k *rsa.PrivateKey) error {
	certPath := fmt.Sprintf("%s.crt", email)
	certFile, err := ensureFile(path, certPath)
	if err != nil {
		return fmt.Errorf("unable to save client certificate: %s", err)
	}
	if err := saveCert(certFile, c); err != nil {
		return err
	}

	keyPath := fmt.Sprintf("%s.key", email)
	keyFile, err := ensureFile(path, keyPath)
	if err != nil {
		return fmt.Errorf("unable to save client certificate: %s", err)
	}

	return saveKey(keyFile, k)
}
