package main

import (
	"os"

	"github.com/ardikabs/dpl/internal/cli"
	"github.com/ardikabs/dpl/internal/log"
)

func main() {
	cli := cli.New()
	if err := cli.Execute(); err != nil {
		log.Error(err, "failed to execute command")
		os.Exit(1)
	}
}
