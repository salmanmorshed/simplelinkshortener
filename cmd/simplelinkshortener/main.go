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

	"github.com/salmanmorshed/simplelinkshortener/internal"
	cliHandlers "github.com/salmanmorshed/simplelinkshortener/internal/cli"
)

func main() {
	CLIApp := &cli.App{
		Usage:     "Create a personal link shortening service",
		ArgsUsage: " ",
		Version:   internal.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "config.yaml",
				Usage: "path to config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:     "init",
				Usage:    "Initialize a config file",
				Category: "Configuration",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					return cliHandlers.InitializeConfigFile(cfgPath)
				},
			},
			{
				Name:     "start",
				Usage:    "Start the web server",
				Category: "Server",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					return cliHandlers.StartServer(c.Context, cfgPath)
				},
			},
			{
				Name:     "useradd",
				Usage:    "Add a new user",
				Category: "User management",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					username := c.Args().First()
					return cliHandlers.AddUser(c.Context, cfgPath, username)
				},
			},
			{
				Name:     "usermod",
				Usage:    "Modify username or password",
				Category: "User management",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					username := c.Args().First()
					return cliHandlers.ModifyUser(c.Context, cfgPath, username)
				},
			},
			{
				Name:     "userdel",
				Usage:    "Delete a user",
				Category: "User management",
				Action: func(c *cli.Context) error {
					cfgPath := c.Value("config").(string)
					username := c.Args().First()
					return cliHandlers.DeleteUser(c.Context, cfgPath, username)
				},
			},
		},
		HideHelpCommand: true,
	}

	var wg sync.WaitGroup
	ctx := context.WithValue(context.Background(), internal.CtxKey("wg"), &wg)
	ctx, cancel := context.WithCancel(ctx)

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, syscall.SIGINT)
	go func() {
		<-sigintCh
		fmt.Println("\nCleaning up...")
		cancel()
		wg.Wait()
		os.Exit(0)
	}()

	err := CLIApp.RunContext(ctx, os.Args)
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, cliHandlers.ErrAborted) {
			os.Exit(1)
		}
	}
}
