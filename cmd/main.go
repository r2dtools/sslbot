package main

import (
	"github.com/r2dtools/sslbot/cmd/cli"
	"github.com/r2dtools/sslbot/config"
)

var Version string

func main() {
	config.Version = Version

	if err := cli.Create().Execute(); err != nil {
		panic(err)
	}
}
