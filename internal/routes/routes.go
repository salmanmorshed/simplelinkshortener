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

func GetRouter(conf *config.AppConfig, db *gorm.DB, codec *intstrcodec.CodecConfig) *gin.Engine {
	if !conf.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	if conf.Server.UseCORS {
		router.Use(CORSMiddleware())
	}

	router.GET("/", func(c *gin.Context) {
		if conf.HomeRedirect == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Redirect(http.StatusFound, conf.HomeRedirect)
	})

	router.GET("/:id", func(c *gin.Context) {
		encodedID := c.Param("id")
		if encodedID == "" || encodedID == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.StrToInt(encodedID)

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
	})

	private := router.Group("/private", BasicAuthMiddleware(db))

	private.GET("/web", ServeEmbeddedFile("web/index.html", "text/html"))

	private.GET("/api/links", func(c *gin.Context) {
		user := c.MustGet("user").(*database.User)

		totalLinkCount := user.GetLinkCount(db)

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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page value"})
			return
		}
		if offset < 0 || offset > totalLinkCount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset value out of bounds"})
			return
		}

		links := make([]database.Link, limit)
		if err := user.FetchLinks(db, &links, limit, offset, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
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
	})

	private.GET("/api/links/:id", func(c *gin.Context) {
		encodedID := c.Param("id")
		if encodedID == "" || encodedID == "favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		decodedID := codec.StrToInt(encodedID)

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
	})

	private.POST("/api/links", func(c *gin.Context) {
		user := c.MustGet("user").(*database.User)

		var data struct {
			URL string `json:"url" form:"url"`
		}
		if err := c.ShouldBind(&data); err != nil {
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

		c.JSON(http.StatusCreated, gin.H{
			"short_url": fmt.Sprintf("%s/%s", utils.GetBaseUrl(conf), codec.IntToStr(int(link.ID))),
		})
	})

	return router
}
