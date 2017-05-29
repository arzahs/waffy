package main

import (
	"log"

	"github.com/unerror/waffy/pkg/cmd/waffyd"
)

func main() {
	if err := waffyd.Start(); err != nil {
		log.Fatal("unable to start waffyd")
	}
}
