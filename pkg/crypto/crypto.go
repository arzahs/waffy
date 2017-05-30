package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

const (
	// DefaultExpiryTime is the default time for CertificatesV
	DefaultExpiryTime = time.Hour * 24 * 365 * 2 // 2 year

	hostKeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement
)

func NewCertificate(
	ca *x509.Certificate,
	signer crypto.PrivateKey,
	signee crypto.PrivateKey,
	server bool,
	commonName string,
	alt ...string,
) (*x509.Certificate, error) {
	rsaKey, subjectId, err := keyAndSubjectId(signee)
	if err != nil {
		return nil, err
	}

	serialLim := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLim)
	if err != nil {
		return nil, fmt.Errorf("unable to generate certificate serial")
	}

	var template = x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(DefaultExpiryTime),

		KeyUsage: hostKeyUsage,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		SubjectKeyId: subjectId,
	}

	if server {
		template.DNSNames = alt
	} else {
		emails := []string{commonName}
		template.EmailAddresses = append(emails, alt...)
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, ca, &rsaKey.PublicKey, signer)
	if err != nil {
		return nil, fmt.Errorf("unable to create server cert: %s", err)
	}

	return x509.ParseCertificate(cert)
}

func NewCertificateAuthority(bits int) (*x509.Certificate, crypto.PrivateKey, error) {
	privKey, err := NewPrivateKey(bits)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create private key: %s", err)
	}

	ca, err := newCertificateAuthority(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create certificate authority: %s", err)
	}

	return ca, privKey, nil
}
