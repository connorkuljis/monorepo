package main

import (
	"log"
	"os"

	"github.com/connorkuljis/monorepo/ssg/internal/command"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			&command.GenerateCommand,
			&command.ServeCommand,
			&command.NewPostCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
