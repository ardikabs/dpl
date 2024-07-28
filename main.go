package main

import (
	"os"

	"github.com/ardikabs/dpl/internal/cli"
	"github.com/go-logr/logr"
)

func main() {
	log := logr.FromSlogHandler(defaultLogHandler)
	cli := cli.New(log)
	if err := cli.Execute(); err != nil {
		log.Error(err, "failed to execute command")
		os.Exit(1)
	}
}
