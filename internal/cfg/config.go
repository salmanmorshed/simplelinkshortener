package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Version = "devel"

type Config struct {
	URLPrefix    string `yaml:"url_prefix,omitempty"`
	HomeRedirect string `yaml:"home_redirect,omitempty"`

	Codec struct {
		Alphabet  string `yaml:"alphabet"`
		BlockSize int    `yaml:"block_size"`
	} `yaml:"codec"`

	Database struct {
		Type      string            `yaml:"type"`
		Host      string            `yaml:"host,omitempty"`
		Port      uint16            `yaml:"port,omitempty"`
		Username  string            `yaml:"username,omitempty"`
		Password  string            `yaml:"password,omitempty"`
		Name      string            `yaml:"name,omitempty"`
		ExtraArgs map[string]string `yaml:"extra_args,omitempty"`
	} `yaml:"database"`

	Server struct {
		Host string `yaml:"host"`
		Port uint16 `yaml:"port"`

		UseTLS         bool   `yaml:"use_tls"`
		TLSCertificate string `yaml:"tls_certificate,omitempty"`
		TLSPrivateKey  string `yaml:"tls_private_key,omitempty"`

		UseCache      bool `yaml:"use_cache,omitempty"`
		CacheCapacity uint `yaml:"cache_capacity,omitempty"`

		UseCORS     bool     `yaml:"use_cors,omitempty"`
		CORSOrigins []string `yaml:"cors_origins,omitempty"`
	} `yaml:"server"`
}

func LoadConfigFromFile(configPath string) (*Config, error) {
	var err error

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
	if err = decoder.Decode(&conf); err != nil {
		return nil, err
	}

	if err = validateConfigValues(conf); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return conf, nil
}

func WriteConfigToFile(configPath string, conf *Config) error {
	var err error

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
