package crypto

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"time"
)

const (
	// PEMRSAType RSA PEM Block Type (RSA) private keys
	PEMRSAType = "RSA PRIVATE KEY"

	// PEMCertificateType RSA PEM Block Type for certificates
	PEMCertificateType = "CERTIFICATE"

	caKeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
)

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
			Type:  PEMRSAType,
			Bytes: x509.MarshalPKCS1PrivateKey(data),
		}
	case *x509.Certificate:
		block = pem.Block{
			Type:  PEMCertificateType,
			Bytes: data.Raw,
		}
	}

	return pem.EncodeToMemory(&block)
}

// EncodePEMWriter encodes a PEM Block to a file, given the keydata given
func EncodePEMWriter(keydata interface{}, f io.Writer) error {
	pBytes := EncodePEM(keydata)
	_, err := f.Write(pBytes)
	if err != nil {
		return fmt.Errorf("unable to write PEM data: %s", err)
	}

	return nil
}

// DecodePEMReader decodes a PEM Block from a file and interprets it as either an *x509.Certificate
// or an *rsa.PrivateKey
func DecodePEMReader(f io.ReadCloser) (dec interface{}, err error) {
	defer func() {
		err = f.Close()
	}()

	buf := bytes.NewBuffer([]byte{})
	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read PEM block: %s", err)
	}

	if buf.Len() == 0 {
		return nil, fmt.Errorf("cannot decode empty file")
	}

	return DecodePEM(buf.Bytes())
}

// DecodePEM decodes a PEM block from the given bytes and interprets it as either an
// *x509.Certificate or an *rsa.PrivateKey
func DecodePEM(asn1Data []byte) (interface{}, error) {
	block, rest := pem.Decode(asn1Data)
	if len(rest) > 0 {
		return nil, fmt.Errorf("additional certificate data decoded in PEM block")
	}

	switch block.Type {
	case PEMRSAType:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case PEMCertificateType:
		return x509.ParseCertificate(block.Bytes)
	}

	return nil, fmt.Errorf("unrecognized PEM type")
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
		NotBefore:    time.Now().Truncate(24 * time.Hour),
		NotAfter:     time.Now().Add(DefaultExpiryTime),

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
