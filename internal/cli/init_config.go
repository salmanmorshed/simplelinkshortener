package cli

import (
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
)

func InitializeConfigFile(cfgPath string) error {
	var (
		err  error
		conf cfg.Config
	)

	prompt1 := promptui.Select{
		Label: "Choose database type",
		Items: []string{"sqlite3", "postgresql"},
	}
	_, conf.Database.Type, err = prompt1.Run()
	if err != nil {
		return ErrAborted
	}

	if conf.Database.Type == "sqlite3" {
		prompt2 := promptui.Prompt{
			Label:     "Path to sqlite file",
			Default:   "db.sqlite3",
			AllowEdit: true,
		}
		conf.Database.Name, err = prompt2.Run()
		if err != nil {
			return ErrAborted
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
			return ErrAborted
		}

		var (
			PortStr string
			PortVal uint64
		)
		prompt3 := promptui.Prompt{
			Label:     "Port",
			Default:   "5432",
			AllowEdit: true,
		}
		PortStr, err = prompt3.Run()
		if err != nil {
			return ErrAborted
		}
		if PortVal, err = strconv.ParseUint(PortStr, 10, 16); err != nil {
			return err
		}
		conf.Database.Port = uint16(PortVal)

		prompt4 := promptui.Prompt{
			Label:     "Username",
			Default:   "postgres",
			AllowEdit: true,
		}
		conf.Database.Username, err = prompt4.Run()
		if err != nil {
			return ErrAborted
		}

		prompt5 := promptui.Prompt{
			Label:   "Password",
			Default: "",
			Mask:    '*',
		}
		conf.Database.Password, err = prompt5.Run()
		if err != nil {
			return ErrAborted
		}

		prompt6 := promptui.Prompt{
			Label:     "Database",
			Default:   "shortener",
			AllowEdit: true,
		}
		conf.Database.Name, err = prompt6.Run()
		if err != nil {
			return ErrAborted
		}

		conf.Database.ExtraArgs = map[string]string{
			"sslmode":  "prefer",
			"timezone": "UTC",
		}
	} else {
		return fmt.Errorf("unsupported database type: %s", conf.Database.Type)
	}

	var useReverseProxy, useTLS, domain string

	prompt7 := promptui.Select{
		Label: "Will it run behind a reverse proxy?",
		Items: []string{"no", "yes"},
	}
	_, useReverseProxy, err = prompt7.Run()
	if err != nil {
		return ErrAborted
	}

	prompt8 := promptui.Select{
		Label: "Use TLS?",
		Items: []string{"no", "yes"},
	}
	_, useTLS, err = prompt8.Run()
	if err != nil {
		return ErrAborted
	}

	prompt9 := promptui.Prompt{
		Label:     "Domain",
		Default:   "example.com",
		AllowEdit: true,
	}
	domain, err = prompt9.Run()
	if err != nil {
		return ErrAborted
	}

	if useReverseProxy == "no" {
		conf.Server.UseTLS = useTLS == "yes"
		conf.Server.Host = domain
		if conf.Server.UseTLS {
			conf.Server.Port = 443

			prompt10 := promptui.Prompt{
				Label:     "Certificate",
				Default:   fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain),
				AllowEdit: true,
			}
			conf.Server.TLSCertificate, err = prompt10.Run()
			if err != nil {
				return ErrAborted
			}

			prompt11 := promptui.Prompt{
				Label:     "PrivateKey",
				Default:   fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", domain),
				AllowEdit: true,
			}
			conf.Server.TLSPrivateKey, err = prompt11.Run()
			if err != nil {
				return ErrAborted
			}
		} else {
			conf.Server.Port = 80
		}
	} else {
		conf.Server.UseTLS = false
		conf.Server.Host = "127.0.0.1"
		conf.Server.Port = 8000
		if conf.Server.UseTLS {
			conf.URLPrefix = fmt.Sprintf("https://%s", domain)
		} else {
			conf.URLPrefix = fmt.Sprintf("http://%s", domain)
		}
	}

	conf.HomeRedirect = "/web"

	conf.Codec.Alphabet = cfg.CreateRandomAlphabet()
	conf.Codec.BlockSize = 20

	if err = cfg.WriteConfigToFile(cfgPath, &conf); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	fmt.Println("Config file generated:", cfgPath)
	return nil
}
