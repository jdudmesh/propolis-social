package store

import (
	"crypto/ecdsa"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"uk.co.dudmesh.propolis/internal/model"
	"uk.co.dudmesh.propolis/pkg/crypt"
)

type publicKeyCache struct {
	db *sqlx.DB
}

func NewPublicKeyCache() (*publicKeyCache, error) {
	db, err := sqlx.Connect("sqlite3", "file:publickeycache.db?mode=memory&cache=shared")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	publicKeyCache := &publicKeyCache{db}
	publicKeyCache.init()

	return publicKeyCache, nil
}

func (s *publicKeyCache) init() {
	// TODO run a gofunc to periodically clear the cache
	s.db.MustExec(`create table if not exists public_key_cache (
		user_id text primary key,
		key text
	)`)
}

func (s *publicKeyCache) Close() error {
	return s.db.Close()
}

func (s *publicKeyCache) Get(userID model.UserID) (*ecdsa.PublicKey, error) {
	var key string
	err := s.db.Get(&key, "SELECT key FROM public_key_cache WHERE user_id = ?", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrorUserNotFound
		}
		return nil, fmt.Errorf("getting public key from cache: %w", err)
	}

	return crypt.DecodePublicKey(key)
}

func (s *publicKeyCache) Set(userID model.UserID, key *ecdsa.PublicKey) error {
	encodedKey, err := crypt.EncodePublicKey(key, string(userID))
	if err != nil {
		return fmt.Errorf("encoding public key: %w", err)
	}
	_, err = s.db.Exec("INSERT INTO public_key_cache (user_id, key) VALUES (?, ?)", userID, encodedKey)
	if err != nil {
		return fmt.Errorf("setting public key in cache: %w", err)
	}
	return nil
}
