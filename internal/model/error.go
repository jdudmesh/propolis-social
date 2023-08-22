package model

import "errors"

var ErrorInvalidUsernameOrPassword = errors.New("invalid username or password")
var ErrorUserNotFound = errors.New("user not found")
var ErrorSenderMismatch = errors.New("sender mismatch")
