package ircd

import "errors"

var (
	errorConnectionNil              = errors.New("connection is nil")
	errorConnectionRemoteAddressNil = errors.New("connection remote address is nil")
)

var (
	errorBadChannelKey = errors.New("bad channel key")
)

var (
	errorBadTargetLength = errors.New("target has no items")
)
