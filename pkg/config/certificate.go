package config

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SaveCA saves the certificate to the filesystem
func SaveCA(certificate *x509.Certificate, key crypto.PrivateKey) error {
	caFile, err := ensureConfigCertFile("ca.cert")
	if err != nil {
		return fmt.Errorf("cannot create cert: %s", err)
	}

	err = saveCert(caFile, certificate)
	if err != nil {
		return fmt.Errorf("unable to save ca certificate: %s", err)
	}

	keyFile, err := ensureConfigCertFile("ca.key")
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
	f, err := ensureConfigCertFile(certFile)
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
	f, err := ensureConfigCertFile(keyFile)
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

func saveKey(f io.Writer, key crypto.PrivateKey) error {
	w := bufio.NewWriter(f)
	switch privKey := key.(type) {
	case *rsa.PrivateKey:

		block := pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		}
		if err := pem.Encode(w, &block); err != nil {
			return fmt.Errorf("unable to format ")
		}
	}
	return w.Flush()
}

func saveCert(f io.Writer, certificate *x509.Certificate) error {
	w := bufio.NewWriter(f)
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	}
	if err := pem.Encode(w, &block); err != nil {
		return err
	}
	return w.Flush()
}

func loadCert(f io.Reader) (*x509.Certificate, error) {
	block, err := decodePEMBlock(f)
	if err != nil {
		return nil, fmt.Errorf("unable to decode certificate PEM block: %s", err)
	}

	return x509.ParseCertificate(block.Bytes)
}

func loadKey(f io.Reader) (crypto.PrivateKey, error) {
	block, err := decodePEMBlock(f)
	if err != nil {
		return nil, fmt.Errorf("unable to decode key PEM block: %s", err)
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func decodePEMBlock(f io.Reader) (*pem.Block, error) {
	buf := bytes.NewBuffer([]byte{})
	_, err := buf.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read certificate file: %s", err)
	}

	if buf.Len() == 0 {
		return nil, fmt.Errorf("cannot decode empty file")
	}

	block, rest := pem.Decode(buf.Bytes())
	if len(rest) > 0 {
		return nil, fmt.Errorf("additional certificate data decoded in PEM block")
	}

	return block, err
}

func ensureConfigCertFile(filename string) (*os.File, error) {
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

	f, err := ensureFile(certPath, filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func loadConfigCertFile(filename string) (*os.File, error) {
	certPath, err := filepath.Abs(cfg.CertPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load certificate path: %s", err)
	}

	path := filepath.Join(certPath, filename)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("the CertPath does not exist")
		}
	}

	return os.Open(path)
}
