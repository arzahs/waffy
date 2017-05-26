package waffyd

import (
	"github.com/spf13/cobra"
)

var RootCmd *cobra.Command

func init() {
	RootCmd = &cobra.Command{
		Use:    "waffyd start",
		Short:  "waffy daemon",
		Long:   "waffy Web Application Firewall",
		Hidden: true,
		Run: func(command *cobra.Command, args []string) {
			if err := command.Help(); err != nil {
				panic("Cannot load help")
			}
		},
	}
}
