package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	cliActions "github.com/salmanmorshed/simplelinkshortener/internal/cli"
)

func main() {
	app := &cli.App{
		Usage:     "Create a personal link shortening service",
		ArgsUsage: " ",
		Version:   cfg.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "config.yml",
				Usage: "path to config file",
			},
		},
		HideHelpCommand: true,
		Commands: []*cli.Command{
			{
				Name:     "init",
				Usage:    "Initialize a config file",
				Category: "Configuration",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					return cliActions.InitializeConfigFile(cfgPath)
				},
			},
			{
				Name:     "start",
				Usage:    "Start the web server",
				Category: "Server",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					return cliActions.StartServer(c.Context, cfgPath)
				},
			},
			{
				Name:     "useradd",
				Usage:    "Add a new user",
				Category: "User management",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					username := c.Args().First()
					return cliActions.AddUser(c.Context, cfgPath, username)
				},
			},
			{
				Name:     "usermod",
				Usage:    "Modify user details",
				Category: "User management",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					username := c.Args().First()
					return cliActions.ModifyUser(c.Context, cfgPath, username)
				},
			},
			{
				Name:     "userdel",
				Usage:    "Delete a user",
				Category: "User management",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					username := c.Args().First()
					return cliActions.DeleteUser(c.Context, cfgPath, username)
				},
			},
		},
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	ctx = context.WithValue(ctx, "ExitWG", &wg)
	ctx, cancel := context.WithCancel(ctx)

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, syscall.SIGINT)

	go func() {
		<-sigintCh
		cancel()
		fmt.Printf("\nWarm shutdown. Please wait...\n")
		wg.Wait()
		os.Exit(0)
	}()

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, cliActions.ErrAborted) {
			os.Exit(1)
		}
	}
}
