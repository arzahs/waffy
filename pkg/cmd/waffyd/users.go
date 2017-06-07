package waffyd

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"log"
	"strconv"

	"gopkg.in/urfave/cli.v1"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/crypto"
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/repository"
	"github.com/unerror/waffy/pkg/services/protos/certificates"
	"github.com/unerror/waffy/pkg/services/protos/users"
	"github.com/unerror/waffy/pkg/cmd"
)

func init() {
	Cmds = append(Cmds, cli.Command{
		Name:     "users",
		Usage:    "Manage waffyd root users",
		Category: "LEADER USER MANAGEMENT",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new root user",
				Flags: append(certificateFlags,
					cli.StringFlag{
						Name:  "full-name",
						Usage: "The user's full name",
					},
					cli.StringFlag{
						Name:  "email",
						Usage: "The user's email address",
					},
					cli.StringFlag{
						Name:  "role",
						Usage: "The user's role",
					},
				),
				Action: cmd.WithConsensusConfig(createUser),
			},
		},
	})
}

func createUser(ctx *cli.Context, db data.Consensus, cfg *config.Config) error {
	// check required fields
	fullName := ctx.String("full-name")
	email := ctx.String("email")
	roleStr := ctx.String("role")

	if fullName == "" || email == "" || roleStr == "" {
		return fmt.Errorf("--full-name, --email and --role are required")
	}

	write := ctx.Bool("overwrite")
	// check if the user already exists
	if _, err := repository.FindUserByEmail(db, email); err != nil {
		write = true
	}

	// create the user if not (or we want to overwrite)
	if write {
		u, cert, key, err := _newUser(fullName, email, roleStr, ctx.Int("key-size"))
		if err != nil {
			return err
		}

		err = repository.CreateCertificate(db, u.Certificate)
		if err != nil {
			return err
		}

		err = config.SaveClientCert(cfg.CertPath, u.Email, cert, key)
		if err != nil {
			return err
		}
		return repository.CreateUser(db, u)
	}

	log.Fatalf("Unable to create user %s since they already exist. --overwrite to force\n", email)
	return nil
}

func _newUser(name, email, roleStr string, keySize int) (*users.User, *x509.Certificate, *rsa.PrivateKey, error) {
	var role users.Role

	roleID, err := strconv.Atoi(roleStr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unknown role: %s", roleStr)
	}
	switch roleID {
	case 0:
		role = users.Role_USER
	case 1:
		role = users.Role_ADMIN
	default:
		return nil, nil, nil, fmt.Errorf("unknown role ID: %d", roleID)
	}

	u := users.User{
		Name:  name,
		Email: email,
		Role:  role,
	}

	ca, caKey, err := config.LoadCA()
	if err != nil {
		return nil, nil, nil, err
	}

	key, err := crypto.NewPrivateKey(keySize)
	if err != nil {
		return nil, nil, nil, err
	}

	cert, err := crypto.NewCertificate(ca, caKey, key, false, u.Email)
	if err != nil {
		return nil, nil, nil, err
	}

	u.Certificate = &certificates.Certificate{
		Subject: &certificates.Subject{
			Email: u.Email,
		},
		SerialNumber: cert.SerialNumber.Bytes(),
		Certificate:  crypto.EncodePEM(cert),
	}

	return &u, cert, key.(*rsa.PrivateKey), nil
}
