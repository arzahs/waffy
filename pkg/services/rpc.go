package services

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/services/protos/nodes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type baseHandler struct {
	s data.Consensus
}

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

	s, err := newRaft()
	if err != nil {
		return err
	}

	handler := baseHandler{
		s: s,
	}

	server := grpc.NewServer(grpc.Creds(creds))

	nodes.RegisterJoinServiceServer(server, &Node{handler})

	return server.Serve(lis)
}

func newRaft() (data.Consensus, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	local, err := data.NewDB(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	return data.NewRaft(cfg.RaftDIR, cfg.RaftListen, local)
}
