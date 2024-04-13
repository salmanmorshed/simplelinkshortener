package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const postgresSetupSQL = `
CREATE TABLE IF NOT EXISTS users (
	username VARCHAR(32) PRIMARY KEY NOT NULL,
	password VARCHAR(255) NOT NULL,
	is_admin BOOLEAN DEFAULT FALSE NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS links (
	id BIGSERIAL PRIMARY KEY NOT NULL,
	url TEXT NOT NULL,
	visits BIGINT DEFAULT 0 NOT NULL,
	created_by VARCHAR(32) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	FOREIGN KEY (created_by) REFERENCES users(username)
);
`

type PostgresStore struct {
	db *sqlx.DB
}

func (s PostgresStore) CreateUser(ctx context.Context, username string, password string) (*User, error) {
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

func (s PostgresStore) RetrieveAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users")
	if err != nil {
		return nil, errors.New("failed to retrieve users")
	}
	return users, nil
}

func (s PostgresStore) RetrieveUser(ctx context.Context, username string) (*User, error) {
	var user User
	q := s.db.Rebind("SELECT * FROM users WHERE username = ?")
	err := s.db.GetContext(ctx, &user, q, username)
	if err != nil {
		return nil, errors.New("failed to retrieve user")
	}
	return &user, nil
}

func (s PostgresStore) UpdateUsername(ctx context.Context, username, newUsername string) error {
	q := s.db.Rebind("UPDATE users SET username = ? WHERE username = ?")
	_, err := s.db.ExecContext(ctx, q, newUsername, username)
	if err != nil {
		return errors.New("failed to update username")
	}
	return nil
}

func (s PostgresStore) UpdatePassword(ctx context.Context, username, newPassword string) error {
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

func (s PostgresStore) ToggleAdmin(ctx context.Context, username string) error {
	q := s.db.Rebind("UPDATE users SET is_admin = NOT is_admin WHERE username = ?")
	_, err := s.db.ExecContext(ctx, q, username)
	if err != nil {
		return errors.New("failed to update admin status")
	}
	return nil
}

func (s PostgresStore) DeleteUser(ctx context.Context, username string) error {
	q := s.db.Rebind("DELETE from users WHERE username = ?")
	_, err := s.db.ExecContext(ctx, q, username)
	if err != nil {
		return errors.New("failed to delete user")
	}
	return nil
}

func (s PostgresStore) CreateLink(ctx context.Context, url, creatorUsername string) (*Link, error) {
	var link Link
	q := s.db.Rebind(`INSERT INTO links (url, created_by) VALUES (?, ?) RETURNING *`)
	err := s.db.GetContext(ctx, &link, q, url, creatorUsername)
	if err != nil {
		return nil, errors.New("failed to create new link")
	}
	return &link, err
}

func (s PostgresStore) RetrieveLink(ctx context.Context, id uint) (*Link, error) {
	var link Link
	err := s.db.GetContext(ctx, &link, s.db.Rebind("SELECT * FROM links WHERE id = ?"), id)
	if err != nil {
		return nil, errors.New("failed to retrieve link")
	}
	return &link, nil
}

func (s PostgresStore) IncrementVisits(ctx context.Context, id uint, count uint) error {
	q := s.db.Rebind("UPDATE links SET visits = visits + ? WHERE id = ?")
	r, err := s.db.ExecContext(ctx, q, count, id)
	if err != nil {
		return errors.New("failed to increment visits")
	}
	if a, err := r.RowsAffected(); err != nil || a != 1 {
		slog.Warn(fmt.Sprintf("failed to increment visits: ID=%d, count=%d, RowsAffected=%d", id, count, a))
	}
	return nil
}

func (s PostgresStore) RetrieveLinkAndBumpVisits(ctx context.Context, id uint) (*Link, error) {
	var link Link
	q := s.db.Rebind("UPDATE links SET visits = visits + 1 WHERE id = ? RETURNING *")
	err := s.db.GetContext(ctx, &link, q, id)
	if err != nil {
		return nil, errors.New("failed to retrieve link")
	}
	return &link, nil
}

func (s PostgresStore) DeleteLink(ctx context.Context, id uint) error {
	_, err := s.db.ExecContext(ctx, s.db.Rebind("DELETE from links WHERE id = ?"), id)
	if err != nil {
		return errors.New("failed to delete link")
	}
	return nil
}

func (s PostgresStore) GetLinkCountForUser(ctx context.Context, username string) uint {
	var count uint
	_ = s.db.GetContext(ctx, &count, s.db.Rebind("SELECT count(*) FROM links where created_by = ?"), username)
	return count
}

func (s PostgresStore) RetrieveLinksForUser(ctx context.Context, username string, limit int, offset int) ([]Link, error) {
	links := make([]Link, limit)
	q := s.db.Rebind("SELECT * FROM links WHERE created_by = ? ORDER BY id DESC LIMIT ? OFFSET ?")
	err := s.db.SelectContext(ctx, &links, q, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links")
	}
	return links, nil
}

func (s PostgresStore) Close() {
	if err := s.db.Close(); err != nil {
		slog.Warn("failed to close database connection")
	}
}
