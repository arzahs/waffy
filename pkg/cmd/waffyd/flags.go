package waffyd

import (
	"strconv"

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
