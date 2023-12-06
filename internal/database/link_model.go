package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	URL    string
	Visits uint
	UserID uint
}

func (link *Link) IncrementVisits(db *gorm.DB) error {
	if db.Model(&link).Update("visits", gorm.Expr("visits + ?", 1)).Error != nil {
		return fmt.Errorf("failed to increment visits")
	}
	return nil
}

func (link *Link) Delete(db *gorm.DB) error {
	if q := db.Delete(&link); q.Error != nil {
		return errors.New("failed to delete link")
	}
	return nil
}

func FetchLinksForUser(db *gorm.DB, user *User, limit int, offset int, inverseOrdering bool) ([]Link, error) {
	var order string
	if inverseOrdering {
		order = "created_at desc"
	} else {
		order = "created_at"
	}
	links := make([]Link, limit)
	err := db.Model(user).Limit(limit).Offset(offset).Order(order).Association("Links").Find(&links)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links")
	}
	return links, nil
}

func GetLinkCountForUser(db *gorm.DB, user *User) int {
	return int(db.Model(*user).Association("Links").Count())
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
