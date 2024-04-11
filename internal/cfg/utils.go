package cfg

import (
	"errors"
	"fmt"
	"math/rand"
)

func CreateRandomAlphabet() string {
	runes := []rune("23456789abcdefghijkmnoprstuvwxyzACDEFHJKLMNPQRTUVWXY")
	for i := len(runes) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func validateConfigValues(conf *Config) error {
	if conf.Database.Type == "postgresql" {
		if conf.Database.Host == "" {
			return errors.New("database: host required")
		}

		if conf.Database.Port == 0 {
			return errors.New("database: port required")
		}

		if conf.Database.Username == "" {
			return errors.New("database: username required")
		}

		if conf.Database.Password == "" {
			return errors.New("database: password required")
		}

		if conf.Database.Name == "" {
			return errors.New("database: name required")
		}
	} else if conf.Database.Type == "sqlite3" {
		if conf.Database.Name == "" {
			return errors.New("database: name (sqlite3 file path) required")
		}
	} else {
		return fmt.Errorf("database: invalid type '%s'", conf.Database.Type)
	}

	if conf.Server.UseTLS {
		if conf.Server.TLSCertificate == "" {
			return errors.New("server: tls_certificate required for use_tls")
		}

		if conf.Server.TLSPrivateKey == "" {
			return errors.New("server: tls_private_key required for use_tls")
		}
	}

	if conf.Server.UseCORS {
		if len(conf.Server.CORSOrigins) < 1 {
			return errors.New("server: at least 1 cors_origin entry required for use_cors")
		}
	}

	if conf.Server.UseCache {
		if conf.Server.CacheCapacity < 1 {
			return errors.New("server: cache_capacity should be a positive integer")
		}
	}

	return nil
}
