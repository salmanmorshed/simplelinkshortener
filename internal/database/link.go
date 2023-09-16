package database

import (
	"fmt"

	"github.com/salmanmorshed/intstrcodec"
	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	URL    string
	Visits uint
	UserID uint
}

func (link *Link) GetShortURL(config *config.AppConfig, codec *intstrcodec.CodecConfig) string {
	encodedID := codec.IntToStr(int(link.ID))
	return fmt.Sprintf("%s/%s", utils.GetBaseUrl(config), encodedID)
}

func (link *Link) IncrementVisits(db *gorm.DB) error {
	if db.Model(&link).Update("visits", gorm.Expr("visits + ?", 1)).Error != nil {
		return fmt.Errorf("failed to increment visits")
	}
	return nil
}

func GetLinkByID(db *gorm.DB, id uint) (*Link, error) {
	var linkRecord Link
	if q := db.First(&linkRecord, id); q.Error != nil {
		return nil, fmt.Errorf("no link found with id=%d", id)
	}
	return &linkRecord, nil
}

func CreateNewLink(db *gorm.DB, url string, user *User) (*Link, error) {
	newLink := Link{URL: url, Visits: 0, UserID: user.ID}
	if q := db.Create(&newLink); q.Error != nil {
		return nil, fmt.Errorf("failed to creare link")
	}
	return &newLink, nil
}
