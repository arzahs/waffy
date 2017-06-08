package config

import (
	"bufio"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"path/filepath"

	wcrypto "github.com/unerror/waffy/pkg/crypto"
)

// SaveCA saves the certificate to the filesystem
func SaveCA(certificate *x509.Certificate, key crypto.PrivateKey) error {
	caFile, err := ensureConfigCertFile("ca.crt", true)
	if err != nil {
		return fmt.Errorf("cannot create cert: %s", err)
	}

	err = saveCert(caFile, certificate)
	if err != nil {
		return fmt.Errorf("unable to save ca certificate: %s", err)
	}

	keyFile, err := ensureConfigCertFile("ca.key", true)
	if err != nil {
		return fmt.Errorf("cannot create key: %s", err)
	}
	err = saveKey(keyFile, key)
	if err != nil {
		return fmt.Errorf("unable to save ca key: %s", err)
	}

	return nil
}

// LoadCA loads the public and private key data about the CA
func LoadCA() (*x509.Certificate, crypto.PrivateKey, error) {
	cf, err := loadConfigCertFile("ca.crt")
	if err != nil {
		return nil, nil, fmt.Errorf("could not load ca certificate")
	}

	cert, err := loadCert(cf)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to load CA certificate: %s", err)
	}

	kf, err := loadConfigCertFile("ca.key")
	if err != nil {
		return cert, nil, fmt.Errorf("could not load ca key: %s", err)
	}

	key, err := loadKey(kf)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to load CA private key: %s", err)
	}

	return cert, key, nil
}

// SaveCert saves the certificate data to the file system
func SaveCert(name string, certificate *x509.Certificate) error {
	certFile := filepath.Join("nodes", name, "node.crt")
	f, err := ensureConfigCertFile(certFile, true)
	if err != nil {
		return fmt.Errorf("unable to create node certificate file: %s", err)
	}

	return saveCert(f, certificate)
}

// LoadCert returns the Certificate from the filesystem
func LoadCert(name string) (*x509.Certificate, error) {
	path := filepath.Join("nodes", name, "node.crt")
	f, err := loadConfigCertFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load key: %s", err)
	}
	return loadCert(f)
}

// SaveKey saves a given private key to the filesystem
func SaveKey(name string, key crypto.PrivateKey) error {
	keyFile := filepath.Join("nodes", name, "node.key")
	f, err := ensureConfigCertFile(keyFile, true)
	if err != nil {
		return fmt.Errorf("unable to create node key file: %s", err)
	}
	return saveKey(f, key)
}

// LoadKey loads the given private key from the filesystem
func LoadKey(name string) (crypto.PrivateKey, error) {
	path := filepath.Join("nodes", name, "node.key")
	f, err := loadConfigCertFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load private key: %s", err)
	}

	return loadKey(f)
}

func saveKey(f io.WriteCloser, key crypto.PrivateKey) (err error) {
	defer func() {
		err = f.Close()
	}()

	w := bufio.NewWriter(f)

	if err := wcrypto.EncodePEMWriter(key, w); err != nil {
		return fmt.Errorf("unable to save privkey: %s", err)
	}

	return w.Flush()
}

func saveCert(f io.WriteCloser, certificate *x509.Certificate) (err error) {
	defer func() {
		err = f.Close()
	}()

	w := bufio.NewWriter(f)

	if err := wcrypto.EncodePEMWriter(certificate, w); err != nil {
		return fmt.Errorf("unable to save pubkey: %s", err)
	}

	return w.Flush()
}

func loadCert(f io.ReadCloser) (c *x509.Certificate, err error) {
	i, err := wcrypto.DecodePEMReader(f)
	if err != nil {
		return nil, err
	}

	return i.(*x509.Certificate), nil
}

func loadKey(f io.ReadCloser) (k crypto.PrivateKey, err error) {
	i, err := wcrypto.DecodePEMReader(f)
	if err != nil {
		return nil, err
	}

	return i.(*rsa.PrivateKey), nil
}

func ensureConfigCertFile(filename string, truncate bool) (*os.File, error) {
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %s", err)
	}

	if _, err := os.Stat(cfg.CertPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.CertPath, 0700); err != nil {
				return nil, err
			}
		}
	}

	certPath, err := filepath.Abs(cfg.CertPath)
	if err != nil {
		return nil, err
	}

	f, err := ensureFile(certPath, filename, truncate)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func loadConfigCertFile(filename string) (*os.File, error) {
	certPath, err := filepath.Abs(cfg.CertPath)
	if err != nil {
		return nil, err
	}
	f, err := ensureFile(certPath, filename, false)
	if err != nil {
		return nil, fmt.Errorf("unable to load file %s", filename)
	}

	return f, nil
}
