package db

import "context"

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
	Close(ctx context.Context)
}