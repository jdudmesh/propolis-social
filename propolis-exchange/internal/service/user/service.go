package user

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/rakutentech/jwk-go/jwk"

	"golang.org/x/crypto/bcrypt"

	"uk.co.dudmesh.propolis/internal/model"
	"uk.co.dudmesh.propolis/internal/store"
	"uk.co.dudmesh.propolis/pkg/crypt"
	"uk.co.dudmesh.propolis/pkg/user"
)

const (
	BufferSize          int    = 4096
	SizeOfIV            int    = 16
	SizeOfHMAC          int    = 32
	SizeOfKey           int    = 32
	PEMPublicKeyHeader  string = "PUBLIC KEY"
	PEMPrivateKeyHeader string = "PRIVATE KEY"
)

type Config interface {
	store.Config
}

type Database interface {
	CreateUser(user *model.User) error
}

type PublicKeyCache interface {
	Get(userID model.UserID) (*ecdsa.PublicKey, error)
	Set(userID model.UserID, key *ecdsa.PublicKey) error
	Close() error
}

type service struct {
	config         Config
	publicKeyCache PublicKeyCache
}

func New(config Config) (*service, error) {
	cache, err := store.NewPublicKeyCache()
	if err != nil {
		return nil, fmt.Errorf("creating public key cache: %w", err)
	}
	return &service{
		config:         config,
		publicKeyCache: cache,
	}, nil
}

func (s *service) Close() error {
	return s.publicKeyCache.Close()
}

func (s *service) Create(params *model.CreateUserParams) (*model.User, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating public/private key pair: %w", err)
	}
	publicKey := privateKey.PublicKey

	userID := model.UserID(user.IDFromPublicKey(&publicKey))

	privateKeyEnc, err := crypt.EncodePrivatekey(privateKey, string(userID), params.Password)
	if err != nil {
		return nil, fmt.Errorf("encrypting private key: %w", err)
	}

	publicKeyEnc, err := crypt.EncodePublicKey(&publicKey, string(userID))
	if err != nil {
		return nil, fmt.Errorf("encoding public key: %w", err)
	}

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
	if err != nil {
		return nil, fmt.Errorf("generating encoded password: %w", err)
	}
	encodedPassword := base64.StdEncoding.EncodeToString(passwordBytes)

	user := &model.User{
		ID:         userID,
		CreatedAt:  time.Now().UTC(),
		Status:     model.UserStatusActive,
		Handle:     params.Handle,
		Email:      params.Email,
		Password:   encodedPassword,
		PublicKey:  publicKeyEnc,
		PrivateKey: privateKeyEnc,
	}

	store, err := store.NewUserStore(user, s.config)
	if err != nil {
		return nil, fmt.Errorf("creating userstore: %w", err)
	}
	defer store.Close()

	s.publicKeyCache.Set(userID, &publicKey)

	return user, nil
}

func (s *service) Fetch(userID model.UserID) (*model.User, error) {
	store, err := store.ForUser(userID, s.config)
	if err != nil {
		return nil, fmt.Errorf("loading userstore: %w", err)
	}
	defer store.Close()

	user, err := store.Fetch()
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	return user, nil
}

func (s *service) PublicKeyFor(address model.UserAddress) (*ecdsa.PublicKey, error) {
	// TODO, handle remote addresses
	userID := model.UserID(address)
	key, err := s.publicKeyCache.Get(userID)
	if err != nil {
		if err == model.ErrorUserNotFound {
			store, err := store.ForUser(userID, s.config)
			if err != nil {
				return nil, fmt.Errorf("loading userstore: %w", err)
			}
			defer store.Close()

			user, err := store.Fetch()
			if err != nil {
				return nil, fmt.Errorf("fetching user: %w", err)
			}
			key, err := publicKeyFromUser(user)
			if err != nil {
				return nil, fmt.Errorf("getting public key from user: %w", err)
			}
			s.publicKeyCache.Set(userID, key)
		}
		return nil, fmt.Errorf("getting public key from cache: %w", err)
	}
	return key, nil
}

func publicKeyFromUser(u *model.User) (*ecdsa.PublicKey, error) {
	return crypt.DecodePublicKey(u.PublicKey)
}

func privateKeyFromUser(user *model.User, password string) (*ecdsa.PrivateKey, error) {
	shaHash := sha256.New()
	shaHash.Write(base58.Decode(string(user.ID)))
	shaHash.Write([]byte(password))
	key := shaHash.Sum(nil)

	parts := strings.Split(user.PrivateKey, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid private key")
	}
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	keyData, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		if err.Error() == "cipher: message authentication failed" {
			return nil, model.ErrorInvalidUsernameOrPassword
		}
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	keySpec, err := jwk.Parse(string(keyData))
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	return keySpec.Key.(*ecdsa.PrivateKey), nil
}
