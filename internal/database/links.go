package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Link struct {
	ID        uint      `db:"id"`
	URL       string    `db:"url"`
	Visits    uint      `db:"visits"`
	CreatedBy string    `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
}

func CreateNewLink(db *sqlx.DB, url string, createdBy *User) (*Link, error) {
	var link Link
	q := db.Rebind(`INSERT INTO links (url, created_by) VALUES (?, ?) RETURNING *`)
	err := db.Get(&link, q, url, createdBy.Username)
	if err != nil {
		return nil, errors.New("failed to create new link")
	}
	return &link, err
}

func RetrieveLink(db *sqlx.DB, id uint) (*Link, error) {
	var link Link
	err := db.Get(&link, db.Rebind("SELECT * FROM links WHERE id = ?"), id)
	if err != nil {
		return nil, errors.New("failed to retrieve link")
	}
	return &link, nil
}

func (link *Link) IncrementVisits(db *sqlx.DB, count uint) error {
	q := db.Rebind("UPDATE links SET visits = visits + ? WHERE id = ? RETURNING *")
	err := db.Get(&link, q, count, link.ID)
	if err != nil {
		return errors.New("failed to increment visits")
	}
	return nil
}

func (link *Link) Delete(db *sqlx.DB) error {
	_, err := db.Exec(db.Rebind("DELETE from links WHERE id = ?"), link.ID)
	if err != nil {
		return errors.New("failed to delete link")
	}
	return nil
}

func GetLinkCountForUser(db *sqlx.DB, user *User) uint {
	var count uint
	_ = db.Get(&count, db.Rebind("SELECT count(*) FROM links where created_by = ?"), user.Username)
	return count
}

func RetrieveLinksForUser(db *sqlx.DB, user *User, limit int, offset int) ([]Link, error) {
	links := make([]Link, limit)
	q := db.Rebind("SELECT * FROM links WHERE created_by = ? ORDER BY id DESC LIMIT ? OFFSET ?")
	err := db.Select(&links, q, user.Username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links")
	}
	return links, nil
}
