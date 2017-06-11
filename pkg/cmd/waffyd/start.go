package waffyd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	"context"

	"github.com/unerror/waffy/pkg/cmd"
	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/crypto"
	"github.com/unerror/waffy/pkg/services"
	"github.com/unerror/waffy/pkg/services/protos/nodes"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	Cmds = append(Cmds, cli.Command{
		Name:   "start",
		Usage:  "Start the waffyd service",
		Flags:  joinFlags,
		Action: cmd.WithConfig(start),
	})
}

func start(ctx *cli.Context, cfg *config.Config) error {
	ca, _, err := config.LoadCA()
	if ca == nil {
		log.Fatalf("unable to load CA cert: %s", err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)

	keypair, err := loadServerKeypair(cfg.RPCName)
	if err != nil {
		log.Fatalf("unable to load server keypair: %s", err)
	}

	a := services.NewAuthentication(pool, keypair)

	if ctx.String("join") != "" {
		conn, err := services.DialClient(ctx.String("join"), a)
		if err != nil {
			log.Fatalf("unable to start gpc client: %s", err)
		}

		client := nodes.NewJoinServiceClient(conn)
		resp, err := client.Join(context.Background(), &nodes.JoinRequest{
			Hostname: cfg.RPCName,
		})
		if err != nil {
			log.Fatalf("unable to join node: %s", err)
		}

		log.Printf("Node %s joined successfully", resp.Hostname)
	}

	log.Printf("starting RPC for %s server on %s", cfg.RPCName, cfg.APIListen)
	if err := services.Serve(cfg.APIListen, a); err != nil {
		log.Fatalf("unable to serve RPC: %s", err)
	}
	return nil
}

func loadServerKeypair(hostname string) (tls.Certificate, error) {
	cert, err := config.LoadCert(hostname)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to load server certificate for %s: %s", hostname, err)
	}

	key, err := config.LoadKey(hostname)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to load server key: %s", err)
	}

	keypair, err := tls.X509KeyPair(
		crypto.EncodePEM(cert),
		crypto.EncodePEM(key),
	)

	return keypair, err
}
