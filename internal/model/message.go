package model

import "time"

type MessageID string

type MessageStatus int

const (
	MessageStatusPending MessageStatus = iota
	MessageStatusSent
	MessageStatusFailed
	MessageStatusFailedPermanent
	MessageStatusDeleted
)

type Message struct {
	Raw         string
	ID          MessageID
	Payload     interface{}
	ContentType string
	Timestamp   time.Time
	Sender      UserID
}
