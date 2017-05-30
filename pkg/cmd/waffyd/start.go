package waffyd

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/crypto"
	"github.com/unerror/waffy/services"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	Cmds = append(Cmds, cli.Command{
		Name:   "start",
		Usage:  "Start the waffyd service",
		Action: start,
	})
}

func start(c *cli.Context) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("unable to load config: %s", err)
	}

	ca, _, err := config.LoadCA()
	if ca == nil {
		log.Fatalf("unable to load CA cert: %s", err)
	}

	pool := crypto.LoadCertificateAuthrotityPool(ca)
	keypair, err := loadServerKeypair(cfg.RPCName)
	if err != nil {
		log.Fatalf("unable to load server keypair: %s", err)
	}

	log.Printf("starting RPC for %s server on %s", cfg.RPCName, cfg.APIListen)
	if err := services.Serve(cfg.APIListen, pool, *keypair); err != nil {
		log.Fatalf("unable to serve RPC: %s", err)
	}
}

func loadServerKeypair(hostname string) (*tls.Certificate, error) {
	certName := fmt.Sprintf("%s.crt", hostname)
	keyName := fmt.Sprintf("%s.key", hostname)

	cert, err := config.LoadCert(certName)
	if err != nil {
		return nil, fmt.Errorf("unable to load server certificate for %s: %s", hostname, err)
	}

	key, err := config.LoadKey(keyName)
	if err != nil {
		return nil, fmt.Errorf("unable to load server key: %s", err)
	}

	keypair, err := tls.X509KeyPair(
		crypto.EncodePEM(cert),
		crypto.EncodePEM(key),
	)

	return &keypair, err
}
