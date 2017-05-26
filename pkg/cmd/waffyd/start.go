package waffyd

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the waffyd",
		Long:  "Start the waffy Web Application Firewall",
		Run:   start,
	})
}

func start(command *cobra.Command, args []string) {

}
