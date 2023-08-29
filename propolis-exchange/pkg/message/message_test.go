package message

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/rakutentech/jwk-go/jwk"
	"github.com/stretchr/testify/assert"

	"uk.co.dudmesh.propolis/pkg/user"
)

func TestMessage(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.Nil(err)

	publicKey := privateKey.PublicKey
	userID := user.IDFromPublicKey(&publicKey)

	payload := map[string]interface{}{
		"data": "hello world",
	}
	m, id, err := New(payload, Address(userID), "application/json", privateKey)
	assert.Nil(err)
	assert.NotNil(m)
	assert.NotNil(id)

	t.Logf("Message: %s", m)
	t.Logf("ID: %s", id)

	m2, err := Parse([]byte(m), func(header *Header) (*ecdsa.PublicKey, error) {
		return &publicKey, nil
	})
	assert.Nil(err)
	assert.NotNil(m2)

	m2data := make(map[string]interface{})
	err = json.Unmarshal(m2.Payload, &m2data)
	assert.Nil(err)

	assert.Equal(m2data["data"], payload["data"])

	jwt2, err := jwt.Parse(m, func(token *jwt.Token) (interface{}, error) {
		return &publicKey, nil
	})
	assert.Nil(err)
	assert.NotNil(jwt2)

	ks := jwk.NewSpec(&publicKey)
	rawJWK, err := ks.ToJWK()
	rawJWK.Use = "sig"
	rawJWK.Alg = "ES256"
	rawJWK.Kid = string(userID)
	rawJWK.Crv = "P-256"
	assert.Nil(err)

	jsonJWK, err := rawJWK.MarshalJSON()
	assert.Nil(err)
	t.Logf("JWK: %s", string(jsonJWK))
}
