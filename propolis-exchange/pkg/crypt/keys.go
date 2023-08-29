package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/rakutentech/jwk-go/jwk"
)

func EncodePrivatekey(privateKey *ecdsa.PrivateKey, userID string, password string) (string, error) {
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

func EncodePublicKey(publicKey *ecdsa.PublicKey, keyID string) (string, error) {
	ks := jwk.NewSpec(publicKey)
	rawJWK, err := ks.ToJWK()
	if err != nil {
		return "", fmt.Errorf("creating JWK: %w", err)
	}

	rawJWK.Use = "sig"
	rawJWK.Alg = "ES256"
	rawJWK.Kid = keyID

	keyData, err := rawJWK.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("marshalling JWK: %w", err)
	}
	return base64.StdEncoding.EncodeToString(keyData), nil
}

func DecodePublicKey(publicKey string) (*ecdsa.PublicKey, error) {
	keyData, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("decoding public key: %w", err)
	}

	keySpec, err := jwk.Parse(string(keyData))
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	return keySpec.Key.(*ecdsa.PublicKey), nil
}
