package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func LinkListHandler(store db.Store, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*db.User)

		totalLinkCount := store.GetLinkCountForUser(c, user.Username)

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

		links, err := store.RetrieveLinksForUser(c, user.Username, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := make([]gin.H, len(links))
		for i, link := range links {
			results[i] = gin.H{
				"slug":       codec.Encode(int(link.ID)),
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

func LinkDetailsHandler(store db.Store, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.Decode(slug)

		link, err := store.RetrieveLink(c, uint(decodedID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		user := c.MustGet("user").(*db.User)
		if user.Username != link.CreatedBy {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"slug":       codec.Encode(int(link.ID)),
			"url":        link.URL,
			"visits":     link.Visits,
			"created_at": link.CreatedAt,
		})
	}
}

func LinkCreateHandler(conf *cfg.Config, store db.Store, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*db.User)

		var data struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || data.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		if !CheckURLValidity(data.URL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is invalid"})
			return
		}

		link, err := store.CreateLink(c, data.URL, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		slug := codec.Encode(int(link.ID))
		if slug == "private" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "please try again"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"short_url": fmt.Sprintf("%s/%s", GetBaseURL(conf), slug),
		})
	}
}

func LinkDeleteHandler(store db.Store, codec *intstrcodec.Codec) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.Decode(slug)

		link, err := store.RetrieveLink(c, uint(decodedID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if err := store.DeleteLink(c, link.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}