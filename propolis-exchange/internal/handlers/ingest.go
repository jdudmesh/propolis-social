package handlers

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
	"uk.co.dudmesh.propolis/internal/model"
	"uk.co.dudmesh.propolis/pkg/message"
)

type UserService interface {
	Create(params *model.CreateUserParams) (*model.User, error)
	PublicKeyFor(address model.UserAddress) (*ecdsa.PublicKey, error)
}

type MessageStrategy interface {
	Do() error
}

type unmarshallerFunc func(*message.Message) (interface{}, error)

func unmarshallPost(message *message.Message) (interface{}, error) {
	var post model.Post
	err := json.Unmarshal(message.Payload, &post)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling post: %w", err)
	}
	return &post, nil
}

var messageStrategies = map[model.ContentType]unmarshallerFunc{
	model.ContentTypePost: unmarshallPost,
}

func UnmarshalMessagePayload(message *message.Message) (MessageStrategy, error) {
	contentType := strings.SplitN(message.ContentType, ";", 2)
	unmarshaller, ok := messageStrategies[model.ContentType(contentType[0])]
	if !ok {
		return nil, fmt.Errorf("unknown content type: %s", message.ContentType)
	}
	return unmarshaller(message)
}

func Ingest(userService UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		body := c.Request().Body
		defer body.Close()

		rawRequest, err := io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("reading request body: %w", err)
		}

		message, err := message.Parse(rawRequest, func(header *message.Header) (*ecdsa.PublicKey, error) {
			return userService.PublicKeyFor(model.UserAddress(header.KeyID))
		})
		if err != nil {
			return fmt.Errorf("parsing message: %w", err)
		}

		return c.JSON(200, message)
	}
}
