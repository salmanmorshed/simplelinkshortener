package utils

import (
	"fmt"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
)

func GetBaseUrl(conf *config.AppConfig) string {
	if conf.Shortener.URLPrefix != "" {
		return conf.Shortener.URLPrefix
	}
	scheme := "http"
	portSuffix := ""
	if conf.Server.UseTls {
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

func StringifyConfigDBExtraArgs(conf *config.AppConfig) string {
	var ret, sep string
	if conf.Database.Type == "postgresql" {
		ret = " "
		sep = " "
	} else if conf.Database.Type == "mysql" {
		ret = "?"
		sep = "&"
	} else {
		return ""
	}
	for k, v := range conf.Database.ExtraArgs {
		ret = fmt.Sprintf("%s%s%s=%s", ret, sep, k, v)
	}
	return ret
}
