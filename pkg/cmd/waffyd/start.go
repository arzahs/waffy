package waffyd

import "gopkg.in/urfave/cli.v1"

func init() {
	Cmds = append(Cmds, cli.Command{
		Name:   "start",
		Usage:  "Start the waffyd service",
		Action: start,
	})
}

func start(c *cli.Context) {

}
