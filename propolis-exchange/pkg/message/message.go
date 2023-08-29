package message

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
)

const (
	AlgorithmES256      = "ES256"
	TypePropolisMessage = "x-propolis-message"
)

type Address string

type Header struct {
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	Version   string `json:"v"`
	Timestamp int64  `json:"ts"`
}

type Message struct {
	Raw         []string
	ID          string
	Header      Header
	ContentType string
	Payload     []byte
	SenderID    Address
}

type PublicKeyFn func(header *Header) (*ecdsa.PublicKey, error)

var (
	ErrorInvalidSignature = errors.New("invalid signature")
	ErrorMissingPayload   = errors.New("missing payload")
	ErrorInvalidMessage   = errors.New("invalid message")
)

func New(payload interface{}, senderAddress Address, messageSubType string, privateKey *ecdsa.PrivateKey) (string, string, error) {
	if payload == nil {
		return "", "", ErrorMissingPayload
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("marshalling payload: %w", err)
	}

	header := &Header{
		KeyID:     string(senderAddress),
		Algorithm: AlgorithmES256,
		Type:      fmt.Sprintf("%s;%s", TypePropolisMessage, messageSubType),
		Version:   "1",
		Timestamp: time.Now().UTC().UnixMilli(),
	}

	message, id, err := sign(header, payloadBytes, string(senderAddress), privateKey)
	if err != nil {
		return "", "", fmt.Errorf("signing message: %w", err)
	}

	return message, id, nil
}

func Parse(data []byte, publicKeyFn PublicKeyFn) (*Message, error) {
	m := &Message{
		Header:  Header{},
		Payload: []byte{},
		Raw:     strings.Split(string(data), "."),
	}

	if len(m.Raw) != 3 {
		return nil, ErrorInvalidMessage
	}

	header, err := decodeSegment(m.Raw[0])
	if err != nil {
		return nil, fmt.Errorf("decoding header: %w", err)
	}
	err = json.Unmarshal(header, &m.Header)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling header: %w", err)
	}

	if m.Header.Algorithm != AlgorithmES256 {
		return nil, fmt.Errorf("unsupported algorithm: %s", m.Header.Algorithm)
	}

	contentTypeParts := strings.SplitN(m.Header.Type, ";", 2)
	if contentTypeParts[0] != TypePropolisMessage {
		return nil, fmt.Errorf("unsupported type: %s", m.Header.Type)
	}
	m.ContentType = contentTypeParts[1]

	if m.Header.Version != "1" {
		return nil, fmt.Errorf("unsupported version: %s", m.Header.Version)
	}

	err = m.verify(publicKeyFn)
	if err != nil {
		return nil, fmt.Errorf("verifying message: %w", err)
	}

	m.Payload, err = decodeSegment(m.Raw[1])
	if err != nil {
		return nil, fmt.Errorf("decoding payload: %w", err)
	}

	return m, nil
}

func sign(header *Header, payloadBytes []byte, senderID string, privateKey *ecdsa.PrivateKey) (string, string, error) {
	sbMsg := strings.Builder{}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", "", fmt.Errorf("marshalling header: %w", err)
	}

	sbMsg.WriteString(encodeSegment(headerBytes))
	sbMsg.WriteString(".")
	sbMsg.WriteString(encodeSegment(payloadBytes))

	shaHash := sha256.New()
	shaHash.Write([]byte(sbMsg.String()))
	hashBytes := shaHash.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashBytes)
	if err != nil {
		return "", "", fmt.Errorf("signing message: %w", err)
	}
	signature := make([]byte, 0, 64)
	signature = append(signature, r.Bytes()...)
	signature = append(signature, s.Bytes()...)

	sbMsg.WriteString(".")
	sbMsg.WriteString(encodeSegment(signature))

	message := sbMsg.String()

	sigHash := sha256.New()
	sigHash.Write(signature)
	sigHashBytes := sigHash.Sum(nil)

	sbID := strings.Builder{}
	sbID.WriteString(base58.Encode(sigHashBytes))
	sbID.WriteString(".")
	sbID.WriteString(senderID)
	id := sbID.String()

	return message, id, nil
}

func (m *Message) verify(publicKeyFn PublicKeyFn) error {
	signingString := strings.Join(m.Raw[:2], ".")

	signature, err := decodeSegment(m.Raw[2])
	if err != nil {
		return fmt.Errorf("decoding signature: %w", err)
	}

	if len(signature) != 64 {
		return ErrorInvalidSignature
	}

	r := new(big.Int).SetBytes(signature[0:32])
	s := new(big.Int).SetBytes(signature[32:64])

	dataHash := sha256.New()
	dataHash.Write([]byte(signingString))
	dataHashBytes := dataHash.Sum(nil)

	pubicKey, err := publicKeyFn(&m.Header)
	if err != nil {
		return fmt.Errorf("getting public key: %w", err)
	}
	ok := ecdsa.Verify(pubicKey, dataHashBytes, r, s)
	if !ok {
		return ErrorInvalidSignature
	}

	sigHash := sha256.New()
	sigHash.Write(signature)
	sigHashBytes := sigHash.Sum(nil)

	sb := strings.Builder{}
	sb.WriteString(base58.Encode(sigHashBytes))
	sb.WriteString(".")
	sb.WriteString(m.Header.KeyID)
	m.ID = sb.String()

	return nil
}

func encodeSegment(seg []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(seg), "=")
}

func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(seg)
}
