package userstore

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"uk.co.dudmesh.propolis/internal/boot"
	"uk.co.dudmesh.propolis/internal/model"
)

type userstore struct {
	userID string
	db     *sqlx.DB
}

func New(user *model.User, config *boot.Config) (*userstore, error) {
	userID := string(user.ID)
	dbName := path.Join(config.DataDirectory, userID+".db")

	isCreating := false
	_, err := os.Stat(dbName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			isCreating = true
		}
	}

	db, err := sqlx.Connect("sqlite3", "file:"+dbName)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	datastore := &userstore{userID, db}
	if isCreating {
		err = datastore.createTables()
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("creating tables: %w", err)
		}
		err = datastore.createUser(user)
		if err != nil {
			return nil, fmt.Errorf("creating user metadata: %w", err)
		}
	}

	return datastore, nil
}

func For(userID model.UserID, config *boot.Config) (*userstore, error) {
	dbName := path.Join(config.DataDirectory, string(userID)+".db")

	_, err := os.Stat(dbName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, model.ErrorUserNotFound
		}
		return nil, fmt.Errorf("checking if database exists: %w", err)
	}

	db, err := sqlx.Connect("sqlite3", "file:"+dbName)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	return &userstore{string(userID), db}, nil
}

func (d *userstore) Close() error {
	return d.db.Close()
}

func (d *userstore) Fetch() (*model.User, error) {
	user := &model.User{}
	err := d.db.Get(user, `select * from user where ID = ?`, d.userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	return user, nil
}

func (d *userstore) createTables() error {
	_, err := d.db.Exec(`create table user(
		ID text not null primary key,
		CreatedAt      DATETIME not null,
		UpdatedAt      DATETIME null,
		LastLoggedInAt DATETIME null,
		LoginAttempts  tinyint not null default 0,
		Status         tinyint not null default 0,
		Handle         text not null,
		Email          text not null,
		Profile        text not null,
		Password       text not null,
		PrivateKey     text not null,
		PublicKey      text not null
	)`)
	if err != nil {
		return fmt.Errorf("creating user table: %w", err)
	}

	_, err = d.db.Exec(`create table outbox(
		ID text not null primary key,
		CreatedAt        DATETIME not null,
		Status           tinyint not null default 0,
		SenderAddress    text not null,
		RecipientAddress text not null,
		Hash             text not null,
		ContentType      text not null,
		Payload          text not null,
		Signature        text not null
	)`)
	if err != nil {
		return fmt.Errorf("creating outbox table: %w", err)
	}

	return nil
}

func (d *userstore) createUser(user *model.User) error {
	res, err := d.db.NamedExec(`insert into user
		(ID, CreatedAt, Handle, Email, Profile, Password, PrivateKey, PublicKey)
		values(:ID, :CreatedAt, :Handle, :Email, :Profile, :Password, :PrivateKey, :PublicKey)`, user)

	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	if rows, err := res.RowsAffected(); rows != 1 {
		return fmt.Errorf("expected 1 row to be affected, got %d", rows)
	} else if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}

	return nil
}

func (d *userstore) PutOutbox(message *model.Message) error {
	panic("TODO")
	// res, err := d.db.NamedExec(`insert into outbox
	// 	(ID, CreatedAt, Status, SenderAddress, RecipientAddress, Hash, ContentType, Payload, Signature)
	// 	values(:ID, :CreatedAt, :Status, :SenderAddress, :RecipientAddress, :Hash, :ContentType, :Payload, :Signature)`, message)

	// if err != nil {
	// 	return fmt.Errorf("inserting outbox entry: %w", err)
	// }
	// if rows, err := res.RowsAffected(); rows != 1 {
	// 	return fmt.Errorf("expected 1 row to be affected, got %d", rows)
	// } else if err != nil {
	// 	return fmt.Errorf("getting rows affected: %w", err)
	// }

	return nil
}
