package database

type UserStore interface {
	CreateUser(username string, password string) (*User, error)
	RetrieveAllUsers() ([]User, error)
	RetrieveUser(username string) (*User, error)
	UpdateUsername(username, newUsername string) error
	UpdatePassword(username, newPassword string) error
	ToggleAdmin(username string) error
	DeleteUser(username string) error
}

type LinkStore interface {
	CreateLink(url, creatorUsername string) (*Link, error)
	RetrieveLink(id uint) (*Link, error)
	IncrementVisits(id uint, count uint) error
	RetrieveLinkAndBumpVisits(id uint) (*Link, error)
	DeleteLink(id uint) error
	GetLinkCountForUser(username string) uint
	RetrieveLinksForUser(username string, limit int, offset int) ([]Link, error)
}

type Store interface {
	UserStore
	LinkStore
	Close()
}
