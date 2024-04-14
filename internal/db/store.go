package db

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
)

type UserStore interface {
	CreateUser(ctx context.Context, username string, password string) (*User, error)
	RetrieveAllUsers(ctx context.Context) ([]User, error)
	RetrieveUser(ctx context.Context, username string) (*User, error)
	UpdateUsername(ctx context.Context, username, newUsername string) error
	UpdatePassword(ctx context.Context, username, newPassword string) error
	ToggleAdmin(ctx context.Context, username string) error
	DeleteUser(ctx context.Context, username string) error
}

type LinkStore interface {
	CreateLink(ctx context.Context, url, creatorUsername string) (*Link, error)
	RetrieveLink(ctx context.Context, id uint) (*Link, error)
	IncrementVisits(ctx context.Context, id uint, count uint) error
	RetrieveLinkAndBumpVisits(ctx context.Context, id uint) (*Link, error)
	DeleteLink(ctx context.Context, id uint) error
	GetLinkCountForUser(ctx context.Context, username string) uint
	RetrieveLinksForUser(ctx context.Context, username string, limit int, offset int) ([]Link, error)
}

type Store interface {
	UserStore
	LinkStore
	Close()
}

func NewStore(conf *cfg.Config) (Store, error) {
	if conf.Database.Type == "postgresql" {
		url := fmt.Sprintf(
			"%s://%s:%s@%s:%d/%s",
			conf.Database.Type,
			conf.Database.Username,
			conf.Database.Password,
			conf.Database.Host,
			conf.Database.Port,
			conf.Database.Name,
		)
		if len(conf.Database.ExtraArgs) > 0 {
			sep := '?'
			for k, v := range conf.Database.ExtraArgs {
				url = fmt.Sprintf("%s%c%s=%s", url, sep, k, v)
				sep = '&'
			}
		}

		db, err := sqlx.Connect("pgx", url)
		if err != nil {
			return nil, err
		}

		db.MustExec(postgresSetupSQL)

		return &PostgresStore{db}, nil
	}

	if conf.Database.Type == "sqlite3" {
		db, err := sqlx.Connect("sqlite3", conf.Database.Name)
		if err != nil {
			return nil, err
		}

		db.MustExec(sqliteSetupSQL)

		return &SqliteStore{PostgresStore{db}}, nil
	}

	return nil, fmt.Errorf("unsupported database type '%s'", conf.Database.Type)
}
