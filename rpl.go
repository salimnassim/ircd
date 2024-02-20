package ircd

import (
	"fmt"
	"strings"
)

type rpl interface {
	format() string
}

// RPL_LUSERCLIENT 251.
// https://modern.ircdocs.horse/#rplluserclient-251
type rplLuserClient struct {
	client string
	// Number of clients.
	users int
	// Number of invisible clients.
	invisible int
	// Number of servers.
	servers int
}

func (r rplLuserClient) format() string {
	return fmt.Sprintf(
		"251 %s :There are %d users and %d invisible on %d servers",
		r.client, r.users, r.invisible, r.servers,
	)
}

// 252 RPL_LUSEROP
// https://modern.ircdocs.horse/#rplluserop-252
type rplLuserOp struct {
	client string
	// Number of operators.
	ops int
}

func (r rplLuserOp) format() string {
	return fmt.Sprintf(
		"252 %s %d :operator(s) online",
		r.client, r.ops,
	)
}

// 254 RPL_LUSERCHANNELS
// https://modern.ircdocs.horse/#rplluserchannels-254
type rplLuserChannels struct {
	client string
	// Number of channels.
	channels int
}

func (r rplLuserChannels) format() string {
	return fmt.Sprintf(
		"254 %s %d :channels formed",
		r.client, r.channels,
	)
}

// 353 RPL_NAMREPLY.
// https://modern.ircdocs.horse/#rplnamreply-353
type rplNamReply struct {
	client string
	// Channel symbol. = public, @ secret, * private.
	symbol string
	// Reply channel.
	channel string
	// List of nicknames on channel prefixed by their mode (e.g. +user).
	nicks []string
}

func (r rplNamReply) format() string {
	nicks := strings.Join(r.nicks, " ")
	return fmt.Sprintf(
		"353 %s %s %s :%s",
		r.client, r.symbol, r.channel, nicks,
	)
}

// 432 ERR_ERRONEUSNICKNAME
// https://modern.ircdocs.horse/#errerroneusnickname-432
type rplErroneusNickname struct {
	client string
	nick   string
}

func (r rplErroneusNickname) format() string {
	return fmt.Sprintf(
		"432 %s %s :Erroneus nickname.",
		r.client, r.nick,
	)
}

type rplNicknameInUse struct {
	client string
	nick   string
}

func (r rplNicknameInUse) format() string {
	return fmt.Sprintf(
		"433 %s %s :Nickname is already in use.",
		r.client, r.nick,
	)
}

// 461 ERR_NEEDMOREPARAMS
// https://modern.ircdocs.horse/#errneedmoreparams-461
type rplNeedMoreParams struct {
	client  string
	command string
}

func (r rplNeedMoreParams) format() string {
	return fmt.Sprintf(
		"461 %s %s :Not enough parameters.",
		r.client, r.command,
	)
}
