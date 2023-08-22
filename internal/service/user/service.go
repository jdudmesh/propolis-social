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
	"io"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/rakutentech/jwk-go/jwk"

	"golang.org/x/crypto/bcrypt"

	"uk.co.dudmesh.propolis/internal/boot"
	"uk.co.dudmesh.propolis/internal/model"
	"uk.co.dudmesh.propolis/internal/userstore"
)

const (
	BufferSize          int    = 4096
	SizeOfIV            int    = 16
	SizeOfHMAC          int    = 32
	SizeOfKey           int    = 32
	PEMPublicKeyHeader  string = "PUBLIC KEY"
	PEMPrivateKeyHeader string = "PRIVATE KEY"
)

type Database interface {
	CreateUser(user *model.User) error
}

type service struct {
	config *boot.Config
}

func New(config *boot.Config) *service {
	return &service{config}
}

func (s *service) Create(params *model.CreateUserParams) (*model.User, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating public/private key pair: %w", err)
	}
	publicKey := privateKey.PublicKey

	userID := UserIDFromPublicKey(&publicKey)

	privateKeyEnc, err := encodePrivatekey(privateKey, userID, params.Password)
	if err != nil {
		return nil, fmt.Errorf("encrypting private key: %w", err)
	}

	publicKeyEnc, err := encodePublicKey(&publicKey, userID)
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

	store, err := userstore.New(user, s.config)
	if err != nil {
		return nil, fmt.Errorf("creating userstore: %w", err)
	}
	defer store.Close()

	return user, nil
}

func (s *service) Fetch(userID model.UserID) (*model.User, error) {
	store, err := userstore.For(userID, s.config)
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
	store, err := userstore.For(userID, s.config)
	if err != nil {
		return nil, fmt.Errorf("loading userstore: %w", err)
	}
	defer store.Close()

	user, err := store.Fetch()
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	return publicKeyFromUser(user)
}

func encodePrivatekey(privateKey *ecdsa.PrivateKey, userID model.UserID, password string) (string, error) {
	ks := jwk.NewSpec(privateKey)
	rawJWK, err := ks.ToJWK()
	if err != nil {
		return "", fmt.Errorf("creating JWK: %w", err)
	}

	rawJWK.Use = "sig"
	rawJWK.Alg = "ES256"
	rawJWK.Kid = string(userID)

	keyData, err := rawJWK.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("marshalling JWK: %w", err)
	}

	shaHash := sha256.New()
	shaHash.Write(base58.Decode(string(userID)))
	shaHash.Write([]byte(password))
	key := shaHash.Sum(nil)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating AES cipher: %w", err)
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("creating AES nonce: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM cipher: %w", err)
	}

	ciphertext := aesgcm.Seal(nil, nonce, keyData, nil)
	sb := strings.Builder{}
	sb.WriteString(base64.StdEncoding.EncodeToString(nonce))
	sb.WriteRune('.')
	sb.WriteString(base64.StdEncoding.EncodeToString(ciphertext))

	return sb.String(), nil
}

func encodePublicKey(publicKey *ecdsa.PublicKey, userID model.UserID) (string, error) {
	ks := jwk.NewSpec(publicKey)
	rawJWK, err := ks.ToJWK()
	if err != nil {
		return "", fmt.Errorf("creating JWK: %w", err)
	}

	rawJWK.Use = "sig"
	rawJWK.Alg = "ES256"
	rawJWK.Kid = string(userID)

	keyData, err := rawJWK.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("marshalling JWK: %w", err)
	}
	return base64.StdEncoding.EncodeToString(keyData), nil
}

func publicKeyFromUser(u *model.User) (*ecdsa.PublicKey, error) {
	keyData, err := base64.StdEncoding.DecodeString(u.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("decoding public key: %w", err)
	}

	keySpec, err := jwk.Parse(string(keyData))
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	return keySpec.Key.(*ecdsa.PublicKey), nil
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

func UserIDFromPublicKey(publicKey *ecdsa.PublicKey) model.UserID {
	shaHash := sha256.New()
	shaHash.Write([]byte(publicKey.X.Bytes()))
	shaHash.Write([]byte(publicKey.Y.Bytes()))
	rawID := shaHash.Sum(nil)
	return model.UserID(base58.Encode(rawID[:]))
}
