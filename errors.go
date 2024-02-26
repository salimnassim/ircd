package ircd

import "errors"

var (
	errorConnectionNil              = errors.New("connection is nil")
	errorConnectionRemoteAddressNil = errors.New("connection remote address is nil")
	errorConnectionLocalAddressNil  = errors.New("connection local address is nil")
)

var (
	errorBadChannelKey = errors.New("bad channel key")
)

var (
	errorCommandNotFound = errors.New("unknown command")
)

var (
	errorBadMaskCharadcter = errors.New("bad mask character")
)

var (
	errorBanMaskDoesNotExist  = errors.New("ban mask does not exist")
	errorBanMaskAlreadyExists = errors.New("ban mask is already defined")
)
