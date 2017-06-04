package waffyd

import (
	"log"
	"strconv"

	"gopkg.in/urfave/cli.v1"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/crypto"
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/repository"
	"github.com/unerror/waffy/pkg/services/protos/certificates"
)

func init() {
	Cmds = append(Cmds, cli.Command{
		Name:     "rsa",
		Usage:    "Manage RSA certificates",
		Category: "CERTIFICATES",
		Subcommands: []cli.Command{
			{
				Name:   "genca",
				Usage:  "Generate CA certificates for RPC",
				Flags:  certificateFlags,
				Action: genca,
			},
			{
				Name:  "gencert",
				Usage: "Generate a server certificaate for RPC",
				Flags: append(certificateFlags, cli.StringFlag{
					Name:  "common-name",
					Usage: "Common Name of the server for the certificate",
				}),
				Action: withConsensus(gencert),
			},
		},
	})
}

func genca(ctx *cli.Context) {
	write := ctx.Bool("overwrite")
	if _, _, err := config.LoadCA(); err != nil {
		write = true
	}

	if write {
		keySize, err := strconv.Atoi(ctx.String("key-size"))
		if err != nil {
			log.Fatalf("unable to load key size: %s", err)
		}

		ca, key, err := crypto.NewCertificateAuthority(keySize)
		if err != nil {
			log.Fatalf("unable to create CA: %s", err)
		}

		if err := config.SaveCA(ca, key); err != nil {
			log.Fatalf("unable to save CA: %s", err)
		}
	} else {
		log.Fatalf("unable to save CA: --overwrite to force and overwrite")
	}
}

func gencert(ctx *cli.Context, db data.Consensus) error {
	write := ctx.Bool("overwrite")
	cn := ctx.String("common-name")
	if cn == "" {
		log.Fatalf("--common-name is required")
	}

	_, err := config.LoadCert(cn)
	if err != nil {
		write = true
	}

	if write {
		keySize, err := strconv.Atoi(ctx.String("key-size"))
		if err != nil {
			log.Fatalf("unable to load key size: %s", err)
		}

		ca, caKey, err := config.LoadCA()
		if err != nil {
			log.Fatalf("unable to load CA: %s", err)
		}

		key, err := crypto.NewPrivateKey(keySize)
		if err != nil {
			log.Fatalf("unable to generate new private key: %s", err)
		}

		cert, err := crypto.NewCertificate(ca, caKey, key, true, cn)
		if err != nil {
			log.Fatalf("unable to generate new certificate: %s", err)
		}

		if err := config.SaveKey(cn, key); err != nil {
			log.Fatalf("unable to save private key: %s", err)
		}
		if err := config.SaveCert(cn, cert); err != nil {
			log.Fatalf("unable to save certificate: %s", err)
		}

		c := &certificates.Certificate{
			Certificate:  crypto.EncodePEM(cert),
			SerialNumber: cert.SerialNumber.Bytes(),
			Subject: &certificates.Subject{
				CommonName: cn,
			},
		}
		return repository.CreateCertificate(db, c)
	}

	log.Fatalf("unable to save certificate for %s, already exists. --overwrite to force", cn)
	return nil
}
