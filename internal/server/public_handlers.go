package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
)

func AppRootHandler(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if conf.HomeRedirect == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Redirect(http.StatusFound, conf.HomeRedirect)
	}
}

func OpenShortUrlHandler(db *sqlx.DB, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.StrToInt(slug)
		if decodedID <= 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		link, err := database.RetrieveLinkAndBumpVisits(db, uint(decodedID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Redirect(http.StatusMovedPermanently, link.URL)
	}
}
