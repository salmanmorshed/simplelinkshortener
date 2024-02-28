package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/salmanmorshed/simplelinkshortener/internal"
	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
)

func inject(handler func(*config.Config, database.Store) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		conf, err := config.LoadConfigFromFile(c.Value("config").(string))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store, err := database.NewPgStore(conf)
		if err != nil {
			return fmt.Errorf("failed to initialize db connection: %w", err)
		}
		defer store.Close()

		return handler(conf, store)
	}
}

func main() {
	app := &cli.App{
		Usage:     "Create a personal link shortening service",
		ArgsUsage: " ",
		Version:   internal.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "config.yaml",
				Usage: "Path to config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:     "init",
				Usage:    "Initialize a config file",
				Category: "Configuration",
				Action:   initConfigFileHandler,
			},
			{
				Name:     "start",
				Usage:    "Start the web server",
				Category: "Server",
				Action:   inject(startServer),
			},
			{
				Name:     "useradd",
				Usage:    "Add a new user",
				Category: "User management",
				Action:   inject(addUser),
			},
			{
				Name:     "usermod",
				Usage:    "Modify username or password",
				Category: "User management",
				Action:   inject(modifyUser),
			},
			{
				Name:     "userdel",
				Usage:    "Delete a user",
				Category: "User management",
				Action:   inject(deleteUser),
			},
		},
		HideHelpCommand: true,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
