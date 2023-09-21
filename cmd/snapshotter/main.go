package main

import (
	"log"

	"github.com/NethermindEth/docker-volumes-snapshotter/cli"
)

func main() {
	cmd := cli.RootCmd()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
