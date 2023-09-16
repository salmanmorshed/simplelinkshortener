package routes

import (
	"math"
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
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	if conf.Shortener.HomeRedirect != "" {
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, conf.Shortener.HomeRedirect)
		})
	}

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

	private.GET("", ServeEmbeddedFile("templates/web.html", "text/html"))

	private.POST("/api/links", func(c *gin.Context) {
		user := c.MustGet("user").(*database.User)

		var data struct {
			URL string `json:"url" form:"url"`
		}
		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		if !utils.CheckURLValidity(data.URL, conf.Shortener.StrictValidator) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is invalid"})
			return
		}

		link, err := database.CreateNewLink(db, data.URL, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"short_url": link.GetShortURL(conf, codec)})
	})

	private.GET("/api/links", func(c *gin.Context) {
		limit := 10
		user := c.MustGet("user").(*database.User)

		totalLinkCount := user.GetLinkCount(db)
		totalPageCount := int(math.Ceil(float64(totalLinkCount) / float64(limit)))

		currentPage, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page value"})
			return
		}
		if currentPage < 1 || currentPage > totalPageCount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "page number out of range"})
			return
		}

		offset := (currentPage - 1) * limit

		links := make([]database.Link, limit)

		if err := user.FetchLinks(db, &links, limit, offset); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		results := make([]gin.H, len(links))
		for i, link := range links {
			results[i] = gin.H{
				"short_url":  link.GetShortURL(conf, codec),
				"url":        link.URL,
				"visits":     link.Visits,
				"created_at": link.CreatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"current_page": currentPage,
			"results":      results,
			"total_pages":  totalPageCount,
		})
	})

	return router
}
