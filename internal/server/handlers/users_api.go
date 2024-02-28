package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
)

func UserListHandler(store database.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := store.RetrieveAllUsers()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
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

func UserCreateHandler(store database.Store) gin.HandlerFunc {
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

		user, err := store.CreateUser(data.Username, data.Password)
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

func UserDetailsEditHandler(store database.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		user, err := store.RetrieveUser(c.Param("username"))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

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
			if err := store.UpdateUsername(user.Username, data.Username); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if data.Password != "" {
			if err := utils.CheckPasswordStrengthValidity(data.Password); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := store.UpdatePassword(user.Username, data.Password); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
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

func UserDeleteHandler(store database.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := store.RetrieveUser(c.Param("username"))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "target user is admin"})
			return
		}

		if err := store.DeleteUser(user.Username); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}
