package main

import (
	"github.com/hw-cs-reps/platform/cmd"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

// VERSION specifies the version of the platform
var VERSION = "0.0.1"

func main() {
	app := &cli.App{
		Name:    "platform",
		Usage:   "a web app to allow class representatives manage information.",
		Version: VERSION,
		Commands: []*cli.Command{
			cmd.CmdStart,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
