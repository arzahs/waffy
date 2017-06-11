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

type Authentication struct {
	caPool  *x509.CertPool
	keypair tls.Certificate
}

func NewAuthentication(p *x509.CertPool, k tls.Certificate) *Authentication {
	return &Authentication{
		caPool:  p,
		keypair: k,
	}
}

func (a *Authentication) Credentials() credentials.TransportCredentials {
	creds := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
		ClientCAs:    a.caPool,
		RootCAs:      a.caPool,
		Certificates: []tls.Certificate{a.keypair},
	}
	creds.BuildNameToCertificate()

	return credentials.NewTLS(creds)
}

// Serve blocks and services the RPC
func Serve(listen string, a *Authentication) error {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("unable to start listener: %s", err)
	}

	s, err := newRaft()
	if err != nil {
		return err
	}

	handler := baseHandler{
		s: s,
	}

	server := grpc.NewServer(grpc.Creds(a.Credentials()))

	nodes.RegisterJoinServiceServer(server, &Node{handler})

	return server.Serve(lis)
}

func DialClient(addr string, a *Authentication) (*grpc.ClientConn, error) {
	return grpc.Dial(addr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(a.Credentials()),
	)
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
