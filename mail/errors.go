package mail

import "errors"

var (
	ErrMsgAlreadySent = errors.New("message is already sent")
)
