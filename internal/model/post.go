package model

import (
	"uk.co.dudmesh.propolis/pkg/message"
)

type PostStatus int

const (
	PostStatusPending PostStatus = iota
	PostStatusSent
	PostStatusFailed
	PostStatusFailedPermanent
	PostStatusDeleted
)

type Post struct {
	message.Message
	Status      PostStatus
	Payload     interface{}
	ContentType string
	Attachments []Attachment
}

type Attachment struct {
	URL         string
	Signature   string
	ContentType string
}
