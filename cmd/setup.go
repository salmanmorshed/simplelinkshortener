package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
	"github.com/urfave/cli/v2"
)

func setupConfig(c *cli.Context) error {
	var conf config.AppConfig
	var err error

	prompt1 := promptui.Select{
		Label: "Choose database type",
		Items: []string{"sqlite", "postgresql", "mysql"},
	}
	_, conf.Database.Type, err = prompt1.Run()
	if err != nil {
		fmt.Println("aborted")
		return nil
	}

	if conf.Database.Type == "sqlite" {
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
	} else {
		var defaultPort, defaultUser string

		if conf.Database.Type == "postgresql" {
			defaultPort = "5432"
			defaultUser = "postgres"
			conf.Database.ExtraArgs = map[string]string{"sslmode": "prefer"}
		} else if conf.Database.Type == "mysql" {
			defaultPort = "3306"
			defaultUser = "root"
			conf.Database.ExtraArgs = map[string]string{"charset": "utf8mb4"}
		} else {
			return fmt.Errorf("unsupported database type: %s", conf.Database.Type)
		}

		fmt.Println("Enter database details")

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
			Default:   defaultPort,
			AllowEdit: true,
		}
		conf.Database.Port, err = prompt3.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		prompt4 := promptui.Prompt{
			Label:     "Username",
			Default:   defaultUser,
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
			Label:     "DB name",
			Default:   "shortener",
			AllowEdit: true,
		}
		conf.Database.Name, err = prompt6.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}
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
		conf.Server.UseTls = useTLS == "yes"
		conf.Server.Host = domain
		if conf.Server.UseTls {
			conf.Server.Port = "443"

			prompt10 := promptui.Prompt{
				Label:     "Certificate",
				Default:   fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain),
				AllowEdit: true,
			}
			conf.Server.TlsFiles.Certificate, err = prompt10.Run()
			if err != nil {
				fmt.Println("aborted")
				return nil
			}

			prompt11 := promptui.Prompt{
				Label:     "PrivateKey",
				Default:   fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", domain),
				AllowEdit: true,
			}
			conf.Server.TlsFiles.PrivateKey, err = prompt11.Run()
			if err != nil {
				fmt.Println("aborted")
				return nil
			}
		} else {
			conf.Server.Port = "80"
		}
	} else {
		conf.Server.UseTls = false
		conf.Server.Host = "127.0.0.1"
		conf.Server.Port = "8000"
		if conf.Server.UseTls {
			conf.URLPrefix = fmt.Sprintf("https://%s", domain)
		} else {
			conf.URLPrefix = fmt.Sprintf("http://%s", domain)
		}
	}

	conf.HomeRedirect = "/private/web"

	conf.Codec.Alphabet = utils.CreateRandomAlphabet()
	conf.Codec.BlockSize = 24
	conf.Codec.MinLength = 5

	configPath := c.Value("config").(string)
	if err2 := config.WriteConfigToFile(configPath, &conf); err2 != nil {
		return fmt.Errorf("failed to initialize config: %v", err2)
	}

	fmt.Println("Config file generated:", configPath)
	return nil
}
