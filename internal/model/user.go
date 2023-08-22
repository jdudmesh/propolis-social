package model

import "time"

type UserID string      // local user id
type UserAddress string // possibly remote user id e.g. 3GFQNuSg3dPqDD1emxv5bqX42oxq@somewhere.com

type UserStatus int

const (
	UserStatusPending UserStatus = iota
	UserStatusActive
	UserStatusLocked
	UserStatusDeleted
)

type CreateUserParams struct {
	Handle   string
	Email    string
	Password string
}

type User struct {
	ID             UserID     `db:"ID"`
	CreatedAt      time.Time  `db:"CreatedAt"`
	UpdatedAt      *time.Time `db:"UpdatedAt"`
	LastLoggedInAt *time.Time `db:"LastLoggedInAt"`
	LoginAttempts  int        `db:"LoginAttempts"`
	Status         UserStatus `db:"Status"`
	Handle         string     `db:"Handle"`
	Email          string     `db:"Email"`
	Profile        string     `db:"Profile"`
	Password       string     `db:"Password"`
	PrivateKey     string     `db:"PrivateKey"`
	PublicKey      string     `db:"PublicKey"`
}
