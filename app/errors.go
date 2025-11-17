package app

import "errors"

var (
	ErrNoRUID   = errors.New("no RUID provided with request")
	ErrRUIDneID = errors.New("id of account doesn't match with RUID")

	ErrMsgAlreadySent        = errors.New("message is already sent")
	ErrEmailAlreadyConfirmed = errors.New("email has already been confirmed")
	ErrWrongMailToken        = errors.New("provided token does not exist")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden          = errors.New("not authorized")

	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)
