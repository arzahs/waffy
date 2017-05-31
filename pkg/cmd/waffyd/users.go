package waffyd

import (
	"fmt"
	"strconv"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/repository"
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
				Flags: []cli.Flag{
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
				},
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

	return repository.CreateUser(db, &u)
}
