package waffyd

import (
	"log"
	"strconv"

	"fmt"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/crypto"
	"gopkg.in/urfave/cli.v1"
)

const (
	DEFAULT_BITS = 4096
)

var certificateFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "key-size",
		Usage: "Key size to use for the CA",
		Value: strconv.Itoa(DEFAULT_BITS),
	},
	cli.BoolFlag{
		Name:  "overwrite",
		Usage: "Overwrite the existing CA data",
	},
}

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
				Action: gencert,
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

func gencert(ctx *cli.Context) {
	write := ctx.Bool("overwrite")
	cn := ctx.String("common-name")
	if cn == "" {
		log.Fatalf("--common-name is required")
	}

	certName := fmt.Sprintf("%s.crt", cn)
	_, err := config.LoadCert(certName)
	if err != nil {
		write = true
	}

	if write {
		keyName := fmt.Sprintf("%s.key", cn)
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

		if err := config.SaveKey(keyName, key); err != nil {
			log.Fatalf("unable to save private key: %s", err)
		}
		if err := config.SaveCert(certName, cert); err != nil {
			log.Fatalf("unable to save certificate: %s", err)
		}
	} else {
		log.Fatalf("unable to save certificate for %s, already exists. --overwrite to force", cn)
	}
}
