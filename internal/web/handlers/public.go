package handlers

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func HomePageHandler(conf *cfg.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if conf.HomeRedirect == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Redirect(http.StatusFound, conf.HomeRedirect)
	}
}

func OpenShortLinkHandler(store db.Store, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.Decode(slug)
		if decodedID <= 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		link, err := store.RetrieveLinkAndBumpVisits(c, uint(decodedID))
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Redirect(http.StatusMovedPermanently, link.URL)
	}
}

func EmbeddedWebpageHandler(efs embed.FS, relPath string) func(*gin.Context) {
	return func(c *gin.Context) {
		content, err := efs.ReadFile(relPath)
		if err != nil {
			log.Println(err)
			if os.IsNotExist(err) {
				c.AbortWithStatus(http.StatusNotFound)
			} else {
				c.String(http.StatusInternalServerError, "Error opening file")
			}
			return
		}

		c.Header("Content-Type", "text/html")
		_, _ = c.Writer.Write(content)
		c.Writer.Flush()
	}
}