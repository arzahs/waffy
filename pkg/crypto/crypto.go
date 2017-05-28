package crypto

import (
	"crypto"
	"crypto/x509"
	"fmt"
)

func NewCertAuthority(bits int) (*x509.Certificate, crypto.PrivateKey, error) {
	privKey, err := PrivateKey(bits)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create private key: %s", err)
	}

	ca, err := CertificateAuthority(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create certificate authority: %s", err)
	}

	return ca, privKey, nil
}
