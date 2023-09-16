package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func RunCLI() {
	app := &cli.App{
		Usage:     "Create a personal link shortening service",
		ArgsUsage: " ",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "./config.yaml",
				Usage: "Path to config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:     "setup",
				Usage:    "Initialize a config file",
				Category: "Configuration",
				Action:   setupConfig,
			},
			{
				Name:     "start",
				Usage:    "Start the web server",
				Category: "Server",
				Action:   startServer,
			},
			{
				Name:     "useradd",
				Usage:    "Add a new user",
				Category: "User management",
				Action:   addUser,
			},
			{
				Name:     "usermod",
				Usage:    "Modify username or password",
				Category: "User management",
				Action:   modifyUser,
			},
			{
				Name:     "userdel",
				Usage:    "Delete a user",
				Category: "User management",
				Action:   deleteUser,
			},
		},
		HideHelpCommand: true,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
