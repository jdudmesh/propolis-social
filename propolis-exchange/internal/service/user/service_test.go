package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"uk.co.dudmesh.propolis/internal/boot"
	"uk.co.dudmesh.propolis/internal/model"
)

func TestCreateUser(t *testing.T) {
	assert := assert.New(t)

	createParams := &model.CreateUserParams{
		Handle:   "testuser",
		Email:    "testuser@testdomain.com",
		Password: "password",
	}

	config, err := boot.Load()
	if err != nil {
		t.Fatalf("failed to load boot config")
	}

	service, err := New(config)
	assert.Nil(err)
	var userID model.UserID

	t.Run("Create", func(t *testing.T) {
		user, err := service.Create(createParams)
		assert.Nil(err)
		assert.NotNil(user)
		if user != nil {
			userID = user.ID
		}
	})

	t.Run("Fetch", func(t *testing.T) {
		user, err := service.Fetch(userID)
		assert.Nil(err)
		assert.NotNil(user)
	})

	t.Run("Fetch Public Key", func(t *testing.T) {
		key, err := service.PublicKeyFor(model.UserAddress(userID))
		assert.Nil(err)
		assert.NotNil(key)
	})

	t.Run("Fetch Private Key", func(t *testing.T) {
		user, err := service.Fetch(userID)
		assert.Nil(err)
		key, err := privateKeyFromUser(user, createParams.Password)
		assert.Nil(err)
		assert.NotNil(key)
	})
}
