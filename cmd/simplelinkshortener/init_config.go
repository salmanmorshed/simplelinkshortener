package main

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
)

func initConfigFileHandler(c *cli.Context) error {
	var err error
	var conf config.Config

	prompt1 := promptui.Select{
		Label: "Choose database type",
		Items: []string{"sqlite3", "postgresql"},
	}
	_, conf.Database.Type, err = prompt1.Run()
	if err != nil {
		fmt.Println("aborted")
		return nil
	}

	if conf.Database.Type == "sqlite3" {
		prompt2 := promptui.Prompt{
			Label:     "Path to sqlite file",
			Default:   "./db.sqlite3",
			AllowEdit: true,
		}
		conf.Database.Name, err = prompt2.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}
	} else if conf.Database.Type == "postgresql" {
		fmt.Println("Enter database connection details")

		prompt2 := promptui.Prompt{
			Label:     "Host",
			Default:   "127.0.0.1",
			AllowEdit: true,
		}
		conf.Database.Host, err = prompt2.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		prompt3 := promptui.Prompt{
			Label:     "Port",
			Default:   "5432",
			AllowEdit: true,
		}
		conf.Database.Port, err = prompt3.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		prompt4 := promptui.Prompt{
			Label:     "Username",
			Default:   "postgres",
			AllowEdit: true,
		}
		conf.Database.Username, err = prompt4.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		prompt5 := promptui.Prompt{
			Label:   "Password",
			Default: "",
			Mask:    '*',
		}
		conf.Database.Password, err = prompt5.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		prompt6 := promptui.Prompt{
			Label:     "Database",
			Default:   "shortener",
			AllowEdit: true,
		}
		conf.Database.Name, err = prompt6.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		conf.Database.ExtraArgs = map[string]string{
			"sslmode":  "prefer",
			"timezone": "UTC",
		}
	} else {
		return fmt.Errorf("unsupported database type: %s", conf.Database.Type)
	}

	var useRP string
	var useTLS string
	var domain string

	prompt7 := promptui.Select{
		Label: "Will it run behind a reverse proxy?",
		Items: []string{"no", "yes"},
	}
	_, useRP, err = prompt7.Run()
	if err != nil {
		fmt.Println("aborted")
		return nil
	}

	prompt8 := promptui.Select{
		Label: "Use TLS?",
		Items: []string{"no", "yes"},
	}
	_, useTLS, err = prompt8.Run()
	if err != nil {
		fmt.Println("aborted")
		return nil
	}

	prompt9 := promptui.Prompt{
		Label:     "Domain",
		Default:   "example.com",
		AllowEdit: true,
	}
	domain, err = prompt9.Run()
	if err != nil {
		fmt.Println("aborted")
		return nil
	}

	if useRP == "no" {
		conf.Server.UseTLS = useTLS == "yes"
		conf.Server.Host = domain
		if conf.Server.UseTLS {
			conf.Server.Port = "443"

			prompt10 := promptui.Prompt{
				Label:     "Certificate",
				Default:   fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain),
				AllowEdit: true,
			}
			conf.Server.TLSFiles.Certificate, err = prompt10.Run()
			if err != nil {
				fmt.Println("aborted")
				return nil
			}

			prompt11 := promptui.Prompt{
				Label:     "PrivateKey",
				Default:   fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", domain),
				AllowEdit: true,
			}
			conf.Server.TLSFiles.PrivateKey, err = prompt11.Run()
			if err != nil {
				fmt.Println("aborted")
				return nil
			}
		} else {
			conf.Server.Port = "80"
		}
	} else {
		conf.Server.UseTLS = false
		conf.Server.Host = "127.0.0.1"
		conf.Server.Port = "8000"
		if conf.Server.UseTLS {
			conf.URLPrefix = fmt.Sprintf("https://%s", domain)
		} else {
			conf.URLPrefix = fmt.Sprintf("http://%s", domain)
		}
	}

	conf.HomeRedirect = "/web"

	conf.Codec.Alphabet = utils.CreateRandomAlphabet()
	conf.Codec.BlockSize = 24
	conf.Codec.MinLength = 5

	configPath := c.Value("config").(string)
	if err2 := config.WriteConfigToFile(configPath, &conf); err2 != nil {
		return fmt.Errorf("failed to initialize config: %w", err2)
	}

	fmt.Println("Config file generated:", configPath)
	return nil
}
