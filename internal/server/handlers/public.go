package handlers

import (
	"embed"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/server/cache"
)

func HomePageHandler(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if conf.HomeRedirect == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Redirect(http.StatusFound, conf.HomeRedirect)
	}
}

func OpenShortLinkHandler(conf *config.Config, store database.Store, codec *intstrcodec.Codec) gin.HandlerFunc {
	if conf.Server.UseCache {
		lc := cache.New(
			conf.Server.CacheConfig.Capacity,
			func(slug string) (*database.Link, error) {
				decodedID := codec.StrToInt(slug)
				if decodedID <= 0 {
					return nil, errors.New("failed to decode slug")
				}

				link, err := store.RetrieveLink(uint(decodedID))
				if err != nil {
					return nil, errors.New("failed to retrieve link")
				}

				return link, nil
			},
			func(link *database.Link, hits uint) error {
				return store.IncrementVisits(link.ID, hits)
			},
			conf.Server.CacheConfig.SyncAfter,
		)

		return func(c *gin.Context) {
			slug := c.Param("slug")
			if slug == "" || slug == "favicon.ico" {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			link, err := lc.Resolve(slug)
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.Redirect(http.StatusMovedPermanently, link.URL)
		}
	} else {
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

			link, err := store.RetrieveLinkAndBumpVisits(uint(decodedID))
			if err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.Redirect(http.StatusMovedPermanently, link.URL)
		}
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
