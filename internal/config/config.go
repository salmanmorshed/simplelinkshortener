package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Version string

type Config struct {
	URLPrefix    string `yaml:"url_prefix,omitempty"`
	HomeRedirect string `yaml:"home_redirect,omitempty"`

	Codec struct {
		Alphabet  string `yaml:"alphabet"`
		BlockSize int    `yaml:"block_size"`
		MinLength int    `yaml:"min_length"`
	} `yaml:"codec"`

	Database struct {
		Type      string            `yaml:"type"`
		Host      string            `yaml:"host,omitempty"`
		Port      string            `yaml:"port,omitempty"`
		Username  string            `yaml:"username,omitempty"`
		Password  string            `yaml:"password,omitempty"`
		Name      string            `yaml:"name,omitempty"`
		ExtraArgs map[string]string `yaml:"extra_args,omitempty"`
	} `yaml:"database"`

	Server struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		UseTLS   bool   `yaml:"use_tls"`
		TLSFiles struct {
			Certificate string `yaml:"certificate,omitempty"`
			PrivateKey  string `yaml:"private_key,omitempty"`
		} `yaml:"tls_files,omitempty"`
		UseCORS bool `yaml:"use_cors,omitempty"`
	} `yaml:"server"`
}

func LoadConfigFromFile(configPath string) (*Config, error) {
	cleanedConfigPath, err := filepath.Abs(filepath.Clean(configPath))
	if err != nil {
		return nil, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("can not open file: %s", cleanedConfigPath)
	}
	defer func() { _ = file.Close() }()

	var conf *Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func WriteConfigToFile(configPath string, conf *Config) error {
	cleanedConfigPath, err := filepath.Abs(filepath.Clean(configPath))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(cleanedConfigPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	encoder := yaml.NewEncoder(file)
	defer func() { _ = encoder.Close() }()
	if err := encoder.Encode(conf); err != nil {
		return err
	}

	return nil
}
