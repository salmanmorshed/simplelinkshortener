package internal

import (
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

type CtxKey string

var Version = "devel"

type App struct {
	Debug bool
	Conf  *cfg.Config
	Codec *intstrcodec.Codec
	Store db.Store
}
