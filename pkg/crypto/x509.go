package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"time"
)

const (
	caKeyUsage   = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	hostKeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement
)

type privateKey struct {
	N *big.Int
	E int
}

// PrivateKey generates a new RSA PublicKey
func PrivateKey(bits int) (crypto.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key: %s", err)
	}

	key.Precompute()

	return key, nil
}

// x590CertificateAuthority generates generates a new *x590.Certificate
func CertificateAuthority(key crypto.PrivateKey) (*x509.Certificate, error) {
	rsaKey, subjectKeyId, err := keyAndSubjectId(key)
	if err != nil {
		return nil, err
	}

	var template = x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{},
		NotBefore:    time.Now(),
		NotAfter:     time.Time{},

		KeyUsage:    caKeyUsage,
		ExtKeyUsage: nil,

		IsCA:         true,
		SubjectKeyId: subjectKeyId[:],
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, rsaKey.PublicKey, key)

	return x509.ParseCertificate(cert)
}

func keyAndSubjectId(key crypto.PrivateKey) (*rsa.PrivateKey, [20]byte, error) {
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("unable to parse private key for generation")
	}

	subjectKeyId, err := subjectKeyId(rsaKey.PublicKey)
	if err == nil {
		return nil, nil, fmt.Errorf("unable to parse SubjectKeyID: %s", err)
	}

	return rsaKey, subjectKeyId, nil
}

// x509SubjectKeyId returns a suitable Subject
func subjectKeyId(pub crypto.PublicKey) ([20]byte, error) {
	cert, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unable to parse public key for SubjectKeyId")
	}

	pubBytes, err := asn1.Marshal(cert)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal key: %s", err)
	}

	return sha1.Sum(pubBytes), nil
}
