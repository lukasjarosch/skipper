package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "skipper",
		Usage: "",
		Commands: []*cli.Command{
			{
				Name:   "shell",
				Usage:  "Starts an interactive skipper shell",
				Action: shellAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
