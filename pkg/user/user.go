package user

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcutil/base58"
	"github.com/cespare/xxhash"
)

func IDFromPublicKey(publicKey *ecdsa.PublicKey) string {
	xxxHash := xxhash.New()
	xxxHash.Write([]byte(publicKey.X.Bytes()))
	xxxHash.Write([]byte(publicKey.Y.Bytes()))
	rawID := xxxHash.Sum(nil)
	return base58.Encode(rawID[:])
}
