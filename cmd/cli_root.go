package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v2"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
)

func inject(handler func(*config.Config, *sqlx.DB) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		conf, err := config.LoadConfigFromFile(c.Value("config").(string))
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		db, err := database.InitDatabaseConnection(conf)
		if err != nil {
			return fmt.Errorf("failed to initialize db connection: %v", err)
		}
		defer func() { _ = db.Close() }()

		return handler(conf, db)
	}
}

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
				Action:   setupConfigFileHandler,
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
