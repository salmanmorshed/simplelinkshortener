package db

import (
	"fmt"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
)

const (
	postgresSetupSQL = `
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
	sqliteSetupSQL = `
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY NOT NULL,
			password TEXT NOT NULL,
			is_admin INTEGER DEFAULT 0 NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			url TEXT NOT NULL,
			visits INTEGER DEFAULT 0 NOT NULL,
			created_by TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			FOREIGN KEY (created_by) REFERENCES users(username)
		);
	`
)

func initPostgresOrSqliteDB(conf *cfg.Config) (*sqlx.DB, error) {
	if conf.Database.Type == "postgresql" {
		db, err := sqlx.Connect("pgx", getPostgresConnectionURL(conf))
		if err != nil {
			return nil, err
		}
		db.MustExec(postgresSetupSQL)
		return db, nil
	} else if conf.Database.Type == "sqlite3" {
		db, err := sqlx.Connect("sqlite3", conf.Database.Name)
		if err != nil {
			return nil, err
		}
		db.MustExec(sqliteSetupSQL)
		return db, nil
	}
	return nil, fmt.Errorf("invalid database type: %s", conf.Database.Type)
}

func getPostgresConnectionURL(conf *cfg.Config) string {
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
	return url
}
