package utils

import (
	"fmt"
	"math/rand"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
)

func CreateRandomAlphabet() string {
	runes := []rune("abcdefghjkmnpqrstuvwxyz23456789")
	for i := len(runes) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func StringifyConfigDBExtraArgs(conf *config.Config) string {
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
