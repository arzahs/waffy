package waffyd

import (
	"log"

	"os"

	"github.com/unerror/waffy/pkg/config"
	"gopkg.in/urfave/cli.v1"
)

// Cmds are the cli.Commands that are Commands on the waffy App
var Cmds []cli.Command

// Start starts the waffyd toold
func Start() error {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("unable to load configuration: %s", err)
	}
	app := cli.NewApp()
	app.Name = "waffyd"
	app.Usage = "waffyd firewall and load balancer"
	app.Version = cfg.Version
	app.Commands = Cmds

	return app.Run(os.Args)
}
