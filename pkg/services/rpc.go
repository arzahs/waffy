package services

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Serve blocks and services the RPC
func Serve(listen string, caPool *x509.CertPool, keypair tls.Certificate) error {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("unable to start listener: %s", err)
	}

	if err != nil {
		return fmt.Errorf("unable to load keypair for listener: %s", err)
	}

	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
		ClientCAs:    caPool,
		Certificates: []tls.Certificate{keypair},
	})

	server := grpc.NewServer(grpc.Creds(creds))

	return server.Serve(lis)
}
