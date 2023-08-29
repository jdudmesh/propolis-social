package handlers

import (
	"github.com/labstack/echo/v4"
	"uk.co.dudmesh.propolis/internal/model"
	"uk.co.dudmesh.propolis/pkg/crypt"
)

func CreateUser(userService UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		params := &model.CreateUserParams{}
		if err := c.Bind(params); err != nil {
			return err
		}
		user, err := userService.Create(params)
		if err != nil {
			return err
		}
		return c.JSON(200, user)
	}
}

func GetPublicKey(userService UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		address := model.UserAddress(c.Param("userAddress"))
		publicKey, err := userService.PublicKeyFor(address)
		if err != nil {
			return err
		}
		keyEncoded, err := crypt.EncodePublicKey(publicKey, string(address))
		if err != nil {
			return err
		}
		return c.JSON(200, keyEncoded)
	}
}
