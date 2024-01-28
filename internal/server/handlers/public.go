package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/caching"
	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
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

func OpenShortLinkHandler(conf *config.Config, db *sqlx.DB, codec *intstrcodec.Codec) gin.HandlerFunc {
	if conf.Server.UseCache {
		cache := caching.NewCache(
			conf.Server.CacheConfig.Capacity,
			func(slug string) (*database.Link, error) {
				decodedID := codec.StrToInt(slug)
				if decodedID <= 0 {
					return nil, errors.New("failed to decode slug")
				}

				link, err := database.RetrieveLink(db, uint(decodedID))
				if err != nil {
					return nil, errors.New("failed to retrieve link")
				}

				return link, nil
			},
			func(link *database.Link, hits uint) error {
				return link.IncrementVisits(db, hits)
			},
			conf.Server.CacheConfig.SyncAfter,
		)

		return func(c *gin.Context) {
			slug := c.Param("slug")
			if slug == "" || slug == "favicon.ico" {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			link, err := cache.Resolve(slug)
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

			link, err := database.RetrieveLinkAndBumpVisits(db, uint(decodedID))
			if err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.Redirect(http.StatusMovedPermanently, link.URL)
		}
	}
}
