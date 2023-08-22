package model

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

func CreateID() string {
	uuid, _ := uuid.NewRandom()
	return base58.Encode(uuid[:])
}
