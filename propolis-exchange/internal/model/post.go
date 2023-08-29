package model

import (
	"uk.co.dudmesh.propolis/pkg/message"
)

type PostID string
type PostStatus int
type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"
	ContentTypeText ContentType = "text/plain"
	ContentTypePost ContentType = "x-propolis-post"
)

const (
	PostStatusPending PostStatus = iota
	PostStatusSent
	PostStatusFailed
	PostStatusFailedPermanent
	PostStatusDeleted
)

type ActionVerb string

const (
	ActionVerbCreate ActionVerb = "create"
	ActionVerbUpdate ActionVerb = "update"
	ActionVerbDelete ActionVerb = "delete"
)

type Action struct {
	message.Message
	ContentType ContentType
	Action      ActionVerb
	Payload     interface{}
}

type Post struct {
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
	InReplyTo   PostID       `json:"inReplyTo,omitempty"`
	Replaces    PostID       `json:"replaces,omitempty"`
	ReplacedBy  PostID       `json:"replacedBy,omitempty"`
	RepostOf    PostID       `json:"repostOf,omitempty"`
}

type Attachment struct {
	URL         string
	Signature   string
	ContentType string
}
