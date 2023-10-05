package main

import (
	"log/slog"
	"os"

	"github.com/NethermindEth/docker-volumes-snapshotter/cli"
)

func main() {
	cmd := cli.RootCmd()
	if err := cmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
