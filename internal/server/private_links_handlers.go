package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
)

func LinkListHandler(db *sqlx.DB, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*database.User)

		totalLinkCount := database.GetLinkCountForUser(db, user)

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit value"})
			return
		}
		if limit < 1 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return
		}

		offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset value"})
			return
		}
		if offset < 0 || uint(offset) > totalLinkCount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset value out of bounds"})
			return
		}

		links, err := database.RetrieveLinksForUser(db, user, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := make([]gin.H, len(links))
		for i, link := range links {
			results[i] = gin.H{
				"slug":       codec.IntToStr(int(link.ID)),
				"url":        link.URL,
				"visits":     link.Visits,
				"created_at": link.CreatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"total":   totalLinkCount,
			"limit":   limit,
			"offset":  offset,
		})
	}
}

func LinkDetailsHandler(db *sqlx.DB, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.StrToInt(slug)

		link, err := database.RetrieveLink(db, uint(decodedID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		user := c.MustGet("user").(*database.User)
		if user.Username != link.CreatedBy {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"slug":       codec.IntToStr(int(link.ID)),
			"url":        link.URL,
			"visits":     link.Visits,
			"created_at": link.CreatedAt,
		})
	}
}

func LinkCreateHandler(conf *config.Config, db *sqlx.DB, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*database.User)

		var data struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || data.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		if !utils.CheckURLValidity(data.URL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is invalid"})
			return
		}

		link, err := database.CreateNewLink(db, data.URL, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		slug := codec.IntToStr(int(link.ID))
		if slug == "private" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "please try again"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"short_url": fmt.Sprintf("%s/%s", utils.GetBaseURL(conf), slug),
		})
	}
}

func LinkDeleteHandler(db *sqlx.DB, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.StrToInt(slug)

		link, err := database.RetrieveLink(db, uint(decodedID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if err := link.Delete(db); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}
