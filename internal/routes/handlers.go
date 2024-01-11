package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"
	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
	"gorm.io/gorm"
)

func AppRootHandler(conf *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if conf.HomeRedirect == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Redirect(http.StatusFound, conf.HomeRedirect)
	}
}

func OpenShortUrlHandler(db *gorm.DB, codec *intstrcodec.CodecConfig) gin.HandlerFunc {
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

		link, err := database.GetLinkByID(db, uint(decodedID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		link.IncrementVisits(db)

		c.Redirect(http.StatusFound, link.URL)
	}
}

func LinkListHandler(db *gorm.DB, codec *intstrcodec.CodecConfig) gin.HandlerFunc {
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
		if offset < 0 || offset > totalLinkCount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset value out of bounds"})
			return
		}

		links, err := database.FetchLinksForUser(db, user, limit, offset, true)
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

func LinkDetailsHandler(db *gorm.DB, codec *intstrcodec.CodecConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.StrToInt(slug)

		link, err := database.GetLinkByID(db, uint(decodedID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		user := c.MustGet("user").(*database.User)
		if user.ID != link.UserID {
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

func LinkCreateHandler(conf *config.AppConfig, db *gorm.DB, codec *intstrcodec.CodecConfig) gin.HandlerFunc {
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
			"short_url": fmt.Sprintf("%s/%s", utils.GetBaseUrl(conf), slug),
		})
	}
}

func LinkDeleteHandler(db *gorm.DB, codec *intstrcodec.CodecConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || slug == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.StrToInt(slug)

		link, err := database.GetLinkByID(db, uint(decodedID))
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

func UserListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := database.GetAllUsers(db)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		results := make([]gin.H, len(users))
		for i, user := range users {
			results[i] = gin.H{
				"id":         user.ID,
				"username":   user.Username,
				"password":   "<secret>",
				"is_admin":   user.IsAdmin,
				"created_at": user.CreatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}

func UserCreateHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.BindJSON(&data); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
			return
		}

		if err := utils.CheckUsernameValidity(data.Username); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := utils.CheckPasswordStrengthValidity(data.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := database.CreateNewUser(db, data.Username, data.Password)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"password":   "<secret>",
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
		})
	}
}

func UserDetailsEditHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		user, err := database.GetUserByID(db, uint(userID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		passwordField := "<secret>"

		if c.Request.Method == "GET" {
			goto respondWithUserDetails
		}

		if err := c.BindJSON(&data); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if data.Username != "" {
			if err := utils.CheckUsernameValidity(data.Username); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := user.UpdateUsername(db, data.Username); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if data.Password != "" {
			if err := utils.CheckPasswordStrengthValidity(data.Password); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := user.UpdatePassword(db, data.Password); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			passwordField = "<updated>"
		}

	respondWithUserDetails:
		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"password":   passwordField,
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
		})
	}
}

func UserDeleteHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		user, err := database.GetUserByID(db, uint(userID))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "target user is admin"})
			return
		}

		if err := user.Delete(db); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}
