package ircd

import "errors"

var (
	errorConnectionNil              = errors.New("connection is nil")
	errorConnectionRemoteAddressNil = errors.New("connection remote address is nil")
)
