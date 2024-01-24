package utils

import (
	"fmt"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
)

func GetBaseURL(conf *config.Config) string {
	if conf.URLPrefix != "" {
		return conf.URLPrefix
	}
	scheme := "http"
	portSuffix := ""
	if conf.Server.UseTLS {
		scheme = "https"
		if conf.Server.Port != "443" {
			portSuffix = fmt.Sprintf(":%s", conf.Server.Port)
		}
	} else {
		if conf.Server.Port != "80" {
			portSuffix = fmt.Sprintf(":%s", conf.Server.Port)
		}
	}
	return fmt.Sprintf("%s://%s%s", scheme, conf.Server.Host, portSuffix)
}
