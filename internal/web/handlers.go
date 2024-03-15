package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/salmanmorshed/simplelinkshortener/internal"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func OpenHomePage(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		if app.Conf.HomeRedirect == "" {
			c.String(http.StatusNotFound, "Page not found")
			return
		}

		c.Redirect(http.StatusFound, app.Conf.HomeRedirect)
	}
}

func OpenShortLink(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		encodedID := c.Param("id")
		if IsBadLinkID(encodedID) {
			c.String(http.StatusNotFound, "Link not found")
			return
		}

		decodedID := app.Codec.Decode(encodedID)
		if decodedID <= 0 {
			c.String(http.StatusNotFound, "Link not found")
			return
		}

		link, err := app.Store.RetrieveLinkAndBumpVisits(c, uint(decodedID))
		if err != nil {
			c.String(http.StatusNotFound, "Link not found")
			return
		}

		c.Redirect(http.StatusMovedPermanently, link.URL)
	}
}

func LinkList(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*db.User)

		totalLinkCount := app.Store.GetLinkCountForUser(c, user.Username)

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid limit value"})
			return
		}
		if limit < 1 || limit > 100 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return
		}

		offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid offset value"})
			return
		}
		if offset < 0 || uint(offset) > totalLinkCount {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "offset value out of bounds"})
			return
		}

		links, err := app.Store.RetrieveLinksForUser(c, user.Username, limit, offset)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := make([]gin.H, len(links))
		for i, link := range links {
			results[i] = gin.H{
				"id":         app.Codec.Encode(int(link.ID)),
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
			"prefix":  GetBaseURL(app.Conf),
		})
	}
}

func LinkCreate(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*db.User)

		var data struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&data); err != nil || data.URL == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		if !CheckURLValidity(data.URL) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "url is invalid"})
			return
		}

		link, err := app.Store.CreateLink(c, data.URL, user.Username)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		encodedID := app.Codec.Encode(int(link.ID))
		if IsBadLinkID(encodedID) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "please try again"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"short_url": fmt.Sprintf("%s/%s", GetBaseURL(app.Conf), encodedID),
		})
	}
}

func LinkDetails(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		encodedID := c.Param("id")
		if encodedID == "" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		decodedID := app.Codec.Decode(encodedID)

		link, err := app.Store.RetrieveLink(c, uint(decodedID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		user := c.MustGet("user").(*db.User)
		if user.Username != link.CreatedBy {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         encodedID,
			"url":        link.URL,
			"visits":     link.Visits,
			"created_at": link.CreatedAt,
		})
	}
}

func LinkDelete(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		encodedID := c.Param("id")
		if encodedID == "" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		decodedID := app.Codec.Decode(encodedID)

		link, err := app.Store.RetrieveLink(c, uint(decodedID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if err := app.Store.DeleteLink(c, link.ID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}

func UserList(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := app.Store.RetrieveAllUsers(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := make([]gin.H, len(users))
		for i, user := range users {
			results[i] = gin.H{
				"username":   user.Username,
				"password":   "<secret>",
				"is_admin":   user.IsAdmin,
				"created_at": user.CreatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}

func UserCreate(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.BindJSON(&data); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
			return
		}

		if err := db.CheckUsernameValidity(data.Username); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.CheckPasswordStrengthValidity(data.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := app.Store.CreateUser(c, data.Username, data.Password)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"username":   user.Username,
			"password":   "<secret>",
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
		})
	}
}

func UserDetailsOrEdit(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		user, err := app.Store.RetrieveUser(c, c.Param("username"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		if c.Request.Method == "GET" {
			goto respondWithUserDetails
		}

		if err := c.BindJSON(&data); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if data.Username != "" {
			if err := db.CheckUsernameValidity(data.Username); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := app.Store.UpdateUsername(c, user.Username, data.Username); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if data.Password != "" {
			if err := db.CheckPasswordStrengthValidity(data.Password); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := app.Store.UpdatePassword(c, user.Username, data.Password); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

	respondWithUserDetails:
		c.JSON(http.StatusOK, gin.H{
			"username":   user.Username,
			"password":   "<secret>",
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
		})
	}
}

func UserDelete(app *internal.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := app.Store.RetrieveUser(c, c.Param("username"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		if user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "target user is admin"})
			return
		}

		if err := app.Store.DeleteUser(c, user.Username); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}
