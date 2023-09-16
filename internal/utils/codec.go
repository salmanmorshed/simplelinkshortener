package utils

import (
	"math/rand"
)

func CreateRandomAlphabet() string {
	runes := []rune("mn6j2c4rv8bpygw95z7hsdaetxuk3fq")
	for i := len(runes) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
