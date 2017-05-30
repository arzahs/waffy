package config

import (
	"bufio"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// SaveCert saves the certificate to the filesystem
func SaveCA(certificate *x509.Certificate, key crypto.PrivateKey) error {
	var err error
	err = saveCert("ca.crt", certificate)
	if err != nil {
		return fmt.Errorf("unable to save ca certificate: %s", err)
	}
	err = saveKey("ca.key", key)
	if err != nil {
		return fmt.Errorf("unable to save ca key: %s", err)
	}

	return nil
}

// LoadCA loads the public and private key data about the CA
func LoadCA() (*x509.Certificate, crypto.PrivateKey, error) {
	cf, err := loadFile("ca.crt")
	if err != nil {
		return nil, nil, fmt.Errorf("could not load ca certificate")
	}

	cert, err := loadCert(cf)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to load CA certificate: %s", err)
	}

	kf, err := loadFile("ca.key")
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
	return saveCert(name, certificate)
}

// LoadCert returns the Certificate from the filesystem
func LoadCert(name string) (*x509.Certificate, error) {
	f, err := loadFile(name)
	if err != nil {
		return nil, fmt.Errorf("unable to load key: %s", err)
	}
	return loadCert(f)
}

// SaveKey saves a given private key to the filesystem
func SaveKey(name string, key crypto.PrivateKey) error {
	return saveKey(name, key)
}

// LoadKey loads the given private key from the filesystem
func LoadKey(name string) (crypto.PrivateKey, error) {
	f, err := loadFile(name)
	if err != nil {
		return nil, fmt.Errorf("unable to load private key: %s", err)
	}

	return loadKey(f)
}

func saveKey(filename string, key crypto.PrivateKey) error {
	f, err := ensureFile(filename)
	if err != nil {
		return fmt.Errorf("cannot save key: %s", err)
	}

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

func saveCert(filename string, certificate *x509.Certificate) error {
	f, err := ensureFile(filename)
	if err != nil {
		return fmt.Errorf("cannot save cert: %s", err)
	}

	w := bufio.NewWriter(f)
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	}
	pem.Encode(w, &block)
	return w.Flush()
}

func loadCert(f *os.File) (*x509.Certificate, error) {
	block, err := decodePEMBlock(f)
	if err != nil {
		return nil, fmt.Errorf("unable to decode certificate PEM block: %s", err)
	}

	return x509.ParseCertificate(block.Bytes)
}

func loadKey(f *os.File) (crypto.PrivateKey, error) {
	block, err := decodePEMBlock(f)
	if err != nil {
		return nil, fmt.Errorf("unable to decode key PEM block: %s", err)
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func decodePEMBlock(f *os.File) (*pem.Block, error) {
	certBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read certificate file: %s", err)
	}

	block, rest := pem.Decode(certBytes)
	if len(rest) > 0 {
		return nil, fmt.Errorf("additional certificate data decoded in PEM block")
	}

	return block, err
}

func ensureFile(filename string) (*os.File, error) {
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %s", err)
	}

	if _, err := os.Stat(cfg.CertPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(cfg.CertPath, 0744)
		}
	}

	path := filepath.Join(cfg.CertPath, filename)
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("unable to create CertPath: %s", err)
	}

	return f, nil
}

func loadFile(filename string) (*os.File, error) {
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
