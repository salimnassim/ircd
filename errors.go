package ircd

import "errors"

var (
	errorConnectionNil              = errors.New("connection is nil")
	errorConnectionRemoteAddressNil = errors.New("connection remote address is nil")
	errorConnectionLocalAddressNil  = errors.New("connection local address is nil")
)

var (
	errorCommandNotFound = errors.New("unknown command")
)

var (
	errorBadMaskCharacter = errors.New("bad mask character")
)

var (
	errorBanMaskDoesNotExist  = errors.New("ban mask does not exist")
	errorBanMaskAlreadyExists = errors.New("ban mask is already defined")
)

var (
	errorParserInputTooLong   = errors.New("message is too long")
	errorParserInputMalformed = errors.New("malformed message")
)

var (
	errNickInUse       = errors.New("nickname is already in use")
	errChannelNotFound = errors.New("channel not found")
)

var (
	errSendQExceeded    = errors.New("SendQ exceeded")
	errExcessFlood      = errors.New("Quit: Excess Flood")
	errConnectionClosed = errors.New("Quit: EOF")

	errKilled = errors.New("killed")
)
