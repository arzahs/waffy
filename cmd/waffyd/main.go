package main

import (
	"log"

	"github.com/enmand/waffy/pkg/cmd/waffyd"
)

func main() {
	if err := waffyd.RootCmd.Execute(); err != nil {
		log.Fatalf("Unable to load root command: %s", err)
	}
}
