package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

const caKeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign

// NewPrivateKey generates a new RSA PublicKey
func NewPrivateKey(bits int) (crypto.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key: %s", err)
	}

	key.Precompute()

	return key, nil
}

// EncodePEM encodes the given certificate or key information as a PEM encoded block
func EncodePEM(keydata interface{}) []byte {
	var block pem.Block

	switch data := keydata.(type) {
	case *rsa.PrivateKey:
		block = pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(data),
		}
	case *x509.Certificate:
		block = pem.Block{
			Type:  "CERTIFICATE",
			Bytes: data.Raw,
		}
	}

	return pem.EncodeToMemory(&block)
}

// x590CertificateAuthority generates generates a new *x590.Certificate
func newCertificateAuthority(key crypto.PrivateKey) (*x509.Certificate, error) {
	rsaKey, subjectID, err := keyAndSubjectID(key)
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
		SubjectKeyId: subjectID[:],
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &rsaKey.PublicKey, rsaKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(cert)
}

func keyAndSubjectID(key crypto.PrivateKey) (*rsa.PrivateKey, []byte, error) {
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("unable to parse private key for generation")
	}

	subjectKeyID, err := getSubjectKeyID(rsaKey.PublicKey)
	if err == nil {
		return nil, nil, fmt.Errorf("unable to parse SubjectKeyID: %s", err)
	}

	return rsaKey, subjectKeyID, nil
}

// getSubjectKeyID returns a suitable Subject
func getSubjectKeyID(pub crypto.PublicKey) ([]byte, error) {
	cert, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unable to parse public key for SubjectKeyId")
	}

	pubBytes, err := asn1.Marshal(cert)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal key: %s", err)
	}

	hash := sha1.Sum(pubBytes)
	return hash[:], nil
}
