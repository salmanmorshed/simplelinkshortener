package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
)

type PgStore struct {
	db *sqlx.DB
}

func NewPgStore(conf *cfg.Config) (Store, error) {
	db, err := initPostgresOrSqliteDB(conf)
	if err != nil {
		return nil, err
	}
	return PgStore{db}, nil
}

func (s PgStore) Close() {
	if err := s.db.Close(); err != nil {
		log.Println("failed to close db connection")
	}
}

func (s PgStore) CreateUser(ctx context.Context, username string, password string) (*User, error) {
	var count uint
	q1 := s.db.Rebind("SELECT count(*) FROM users where username = ?")
	err := s.db.GetContext(ctx, &count, q1, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("%s is already taken", username)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	var user User
	q2 := s.db.Rebind("INSERT INTO users (username, password) VALUES (?, ?) RETURNING *")
	err = s.db.GetContext(ctx, &user, q2, username, string(hashedBytes))
	if err != nil {
		return nil, errors.New("failed to create new user")
	}

	return &user, nil
}

func (s PgStore) RetrieveAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users")
	if err != nil {
		return nil, errors.New("failed to retrieve users")
	}
	return users, nil
}

func (s PgStore) RetrieveUser(ctx context.Context, username string) (*User, error) {
	var user User
	q := s.db.Rebind("SELECT * FROM users WHERE username = ?")
	err := s.db.GetContext(ctx, &user, q, username)
	if err != nil {
		return nil, errors.New("failed to retrieve user")
	}
	return &user, nil
}

func (s PgStore) UpdateUsername(ctx context.Context, username, newUsername string) error {
	q := s.db.Rebind("UPDATE users SET username = ? WHERE username = ?")
	_, err := s.db.ExecContext(ctx, q, newUsername, username)
	if err != nil {
		return errors.New("failed to update username")
	}
	return nil
}

func (s PgStore) UpdatePassword(ctx context.Context, username, newPassword string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	q := s.db.Rebind("UPDATE users SET password = ? WHERE username = ?")
	_, err = s.db.ExecContext(ctx, q, string(hashedBytes), username)
	if err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s PgStore) ToggleAdmin(ctx context.Context, username string) error {
	q := s.db.Rebind("UPDATE users SET is_admin = NOT is_admin WHERE username = ?")
	_, err := s.db.ExecContext(ctx, q, username)
	if err != nil {
		return errors.New("failed to update admin status")
	}
	return nil
}

func (s PgStore) DeleteUser(ctx context.Context, username string) error {
	q := s.db.Rebind("DELETE from users WHERE username = ?")
	_, err := s.db.ExecContext(ctx, q, username)
	if err != nil {
		return errors.New("failed to delete user")
	}
	return nil
}

func (s PgStore) CreateLink(ctx context.Context, url, creatorUsername string) (*Link, error) {
	var link Link
	q := s.db.Rebind(`INSERT INTO links (url, created_by) VALUES (?, ?) RETURNING *`)
	err := s.db.GetContext(ctx, &link, q, url, creatorUsername)
	if err != nil {
		return nil, errors.New("failed to create new link")
	}
	return &link, err
}

func (s PgStore) RetrieveLink(ctx context.Context, id uint) (*Link, error) {
	var link Link
	err := s.db.GetContext(ctx, &link, s.db.Rebind("SELECT * FROM links WHERE id = ?"), id)
	if err != nil {
		return nil, errors.New("failed to retrieve link")
	}
	return &link, nil
}

func (s PgStore) IncrementVisits(ctx context.Context, id uint, count uint) error {
	q := s.db.Rebind("UPDATE links SET visits = visits + ? WHERE id = ?")
	r, err := s.db.ExecContext(ctx, q, count, id)
	if err != nil {
		return errors.New("failed to increment visits")
	}
	if a, err := r.RowsAffected(); err != nil || a != 1 {
		log.Printf("failed to increment visits: ID=%d, count=%d, RowsAffected=%d", id, count, a)
	}
	return nil
}

func (s PgStore) RetrieveLinkAndBumpVisits(ctx context.Context, id uint) (*Link, error) {
	var link Link
	q := s.db.Rebind("UPDATE links SET visits = visits + 1 WHERE id = ? RETURNING *")
	err := s.db.GetContext(ctx, &link, q, id)
	if err != nil {
		return nil, errors.New("failed to retrieve link")
	}
	return &link, nil
}

func (s PgStore) DeleteLink(ctx context.Context, id uint) error {
	_, err := s.db.ExecContext(ctx, s.db.Rebind("DELETE from links WHERE id = ?"), id)
	if err != nil {
		return errors.New("failed to delete link")
	}
	return nil
}

func (s PgStore) GetLinkCountForUser(ctx context.Context, username string) uint {
	var count uint
	_ = s.db.GetContext(ctx, &count, s.db.Rebind("SELECT count(*) FROM links where created_by = ?"), username)
	return count
}

func (s PgStore) RetrieveLinksForUser(ctx context.Context, username string, limit int, offset int) ([]Link, error) {
	links := make([]Link, limit)
	q := s.db.Rebind("SELECT * FROM links WHERE created_by = ? ORDER BY id DESC LIMIT ? OFFSET ?")
	err := s.db.SelectContext(ctx, &links, q, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links")
	}
	return links, nil
}
