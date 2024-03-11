package internal

import (
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

var Version = "devel"

type App struct {
	Conf  *cfg.Config
	Codec *intstrcodec.Codec
	Store db.Store
}