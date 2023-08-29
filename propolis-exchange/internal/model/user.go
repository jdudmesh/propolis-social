package model

import "time"

type UserID string      // local user id e.g. 3GFQNuSg3dPqDD1emxv5bqX42oxq
type UserAddress string // possibly remote user id e.g. 3GFQNuSg3dPqDD1emxv5bqX42oxq@somewhere.com

type UserStatus int

const (
	UserStatusPending UserStatus = iota
	UserStatusActive
	UserStatusLocked
	UserStatusDeleted
)

type CreateUserParams struct {
	Handle   string `json:"handle"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID             UserID     `db:"ID" json:"id"`
	CreatedAt      time.Time  `db:"CreatedAt" json:"createdAt"`
	UpdatedAt      *time.Time `db:"UpdatedAt" json:"updatedAt"`
	LastLoggedInAt *time.Time `db:"LastLoggedInAt" json:"-"`
	LoginAttempts  int        `db:"LoginAttempts" json:"-"`
	Status         UserStatus `db:"Status" json:"status"`
	Handle         string     `db:"Handle" json:"handle"`
	Email          string     `db:"Email" json:"email"`
	Profile        string     `db:"Profile" json:"profile"`
	Password       string     `db:"Password" json:"-"`
	PrivateKey     string     `db:"PrivateKey" json:"-"`
	PublicKey      string     `db:"PublicKey" json:"publicKey"`
}
