package waffyd

import (
	"strconv"

	"gopkg.in/urfave/cli.v1"
)

const (
	// DefaultBits is the default bitsize to use for certificate generation
	DefaultBits = 4096
)

var certificateFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "key-size",
		Usage: "Key size to use for the CA",
		Value: strconv.Itoa(DefaultBits),
	},
	cli.BoolFlag{
		Name:  "overwrite",
		Usage: "Overwrite the existing CA data",
	},
}
