package web

import (
	"fmt"
	"net/url"
	"slices"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
)

var badLinkIDs = []string{"", "api", "web", "favicon.ico"}

func GetBaseURL(conf *cfg.Config) string {
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

func CheckURLValidity(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	return true
}

func IsBadLinkID(encodedID string) bool {
	return slices.Contains(badLinkIDs, encodedID)
}
