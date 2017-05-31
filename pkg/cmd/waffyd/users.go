package waffyd

import (
	"fmt"
	"strconv"

	"crypto/rsa"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/crypto"
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/repository"
	"github.com/unerror/waffy/pkg/services/protos/certificates"
	"github.com/unerror/waffy/pkg/services/protos/users"
	"gopkg.in/urfave/cli.v1"
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
				Action: createUser,
			},
		},
	})
}

func createUser(ctx *cli.Context) error {
	// check required fields
	fullName := ctx.String("full-name")
	email := ctx.String("email")
	roleStr := ctx.String("role")
	var role users.Role

	if fullName == "" || email == "" || roleStr == "" {
		return fmt.Errorf("--full-name, --email and --role are required")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := data.NewDB(cfg.DBPath)
	if err != nil {
		return err
	}

	write := ctx.Bool("overwrite")
	// check if the user already exists
	if _, err := repository.FindUserByEmail(db, email); err != nil {
		write = true
	}

	// create the user if not (or we want to overwrite)
	if write {
		roleId, err := strconv.Atoi(roleStr)
		if err != nil {
			return fmt.Errorf("unknown role: %s", roleStr)
		}
		switch roleId {
		case 0:
			role = users.Role_USER
		case 1:
			role = users.Role_ADMIN
		default:
			return fmt.Errorf("unknown role ID: %d", roleId)
		}

		u := users.User{
			Name:  fullName,
			Email: email,
			Role:  role,
		}

		ca, caKey, err := config.LoadCA()

		key, err := crypto.NewPrivateKey(ctx.Int("key-size"))
		if err != nil {
			return err
		}

		cert, err := crypto.NewCertificate(ca, caKey, key, false, u.Email)
		u.Certificate = &certificates.Certificate{
			Subject: &certificates.Subject{
				Email: u.Email,
			},
			SerialNumber: cert.SerialNumber.Bytes(),
			Certificate:  crypto.EncodePEM(cert),
		}

		err = repository.CreateCertificate(db, u.Certificate)
		if err != nil {
			return err
		}

		err = config.SaveClientCert(u.Email, cert, key.(*rsa.PrivateKey))
		if err != nil {
			return err
		}

		return repository.CreateUser(db, &u)
	} else {
		fmt.Print("Unable to create user %s since they already exist. --overwrite to force\n")
	}

	return nil
}
