package config

import (
	"bufio"
	"crypto/x509"
	"fmt"
	"os"
)

// SaveCert saves the certificate to the filesystem
func SaveCACert(certificate *x509.Certificate) error {
	return saveCert("ca.crt", certificate)
}

func SaveCert(name string, certificate *x509.Certificate) error {
	return SaveCert(name, certificate)
}

func saveCert(filename string, certificate *x509.Certificate) error {
	cfg, err := Load()
	if err != nil {
		return fmt.Errorf("unable to load configu: %s", err)
	}

	if _, err := os.Stat(cfg.CertPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(cfg.CertPath, 0744)
		}
	}

	f, err := os.Create(cfg.CertPath)
	if err != nil {
		return fmt.Errorf("unable to save certificates: %s", err)
	}

	w := bufio.NewWriter(f)
	w.Write(certificate.Raw)
	return w.Flush()
}
