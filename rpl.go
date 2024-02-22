package ircd

import (
	"fmt"
	"strings"
)

type rpl interface {
	format() string
}

// 001 RPL_WELCOME
// https://modern.ircdocs.horse/#rplwelcome-001
type rplWelcome struct {
	client   string
	network  string
	hostname string
}

func (r rplWelcome) format() string {
	return fmt.Sprintf(
		"001 %s :Welcome to the %s Network, %s",
		r.client, r.network, r.hostname,
	)
}

// 002 RPL_YOURHOST
// https://modern.ircdocs.horse/#rplyourhost-002
type rplYourHost struct {
	client     string
	serverName string
	version    string
}

func (r rplYourHost) format() string {
	return fmt.Sprintf(
		"002 %s :Your host is %s, running version %s",
		r.client, r.serverName, r.version,
	)
}

// 221 RPL_UMODEIS
// https://modern.ircdocs.horse/#rplumodeis-221
type rplUModeIs struct {
	client     string
	modestring string
}

func (r rplUModeIs) format() string {
	return fmt.Sprintf(
		"221 %s %s",
		r.client, r.modestring,
	)
}

// 251 RPL_LUSERCLIENT.
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

// 311 RPL_WHOISUSER
// https://modern.ircdocs.horse/#rplwhoisuser-311
type rplWhoisUser struct {
	client   string
	nick     string
	username string
	host     string
	realname string
}

func (r rplWhoisUser) format() string {
	return fmt.Sprintf(
		"311 %s %s %s %s * :%s",
		r.client, r.nick, r.username, r.host, r.realname,
	)
}

// 315 RPL_ENDOFWHO
// https://modern.ircdocs.horse/#rplendofwho-315
type rplEndOfWho struct {
	client string
	mask   string
}

func (r rplEndOfWho) format() string {
	return fmt.Sprintf(
		"315 %s %s :End of WHO list.",
		r.client, r.mask,
	)
}

// 319 RPL_WHOISCHANNELS
// https://modern.ircdocs.horse/#rplwhoischannels-319
type rplWhoisChannels struct {
	client   string
	nick     string
	channels []string
}

func (r rplWhoisChannels) format() string {
	channels := strings.Join(r.channels, " ")
	return fmt.Sprintf(
		"319 %s %s :%s",
		r.client, r.nick, channels,
	)
}

// 324 RPL_CHANNELMODEIS
// https://modern.ircdocs.horse/#rplchannelmodeis-324
type rplChannelModeIs struct {
	client     string
	channel    string
	modestring string
	modeargs   string
}

func (r rplChannelModeIs) format() string {
	return fmt.Sprintf(
		"324 %s %s %s %s",
		r.client, r.channel, r.modestring, r.modeargs,
	)
}

// 331 RPL_NOTOPIC
// https://modern.ircdocs.horse/#rplnotopic-331
type rplNoTopic struct {
	client  string
	channel string
}

func (r rplNoTopic) format() string {
	return fmt.Sprintf(
		"331 %s %s :No topic is set.",
		r.client, r.channel,
	)
}

// 331 RPL_TOPIC
// https://modern.ircdocs.horse/#rpltopic-332
type rplTopic struct {
	client  string
	channel string
	topic   string
}

func (r rplTopic) format() string {
	return fmt.Sprintf(
		"332 %s %s :%s",
		r.client, r.channel, r.topic,
	)
}

// 333 RPL_TOPICWHOTIME
// https://modern.ircdocs.horse/#rpltopicwhotime-333
type rplTopicWhoTime struct {
	client  string
	channel string
	nick    string
	setat   int
}

func (r rplTopicWhoTime) format() string {
	return fmt.Sprintf(
		"333 %s %s %s %d",
		r.client, r.channel, r.nick, r.setat,
	)
}

// 352 RPL_WHOREPLY
// https://modern.ircdocs.horse/#rplwhoreply-352
type rplWhoReply struct {
	client   string
	channel  string
	username string
	host     string
	server   string
	nick     string
	flags    string
	hopcount int
	realname string
}

func (r rplWhoReply) format() string {
	return fmt.Sprintf(
		"352 %s %s %s %s %s %s %s :%d %s",
		r.client, r.channel, r.username, r.host, r.server, r.nick, r.flags, r.hopcount, r.realname,
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

// 366 RPL_ENDOFNAMES
type rplEndOfNames struct {
	client  string
	channel string
}

func (r rplEndOfNames) format() string {
	return fmt.Sprintf(
		"366 %s %s :End of /NAMES list.",
		r.client, r.channel,
	)
}

// 372 RPL_MOTD
// https://modern.ircdocs.horse/#rplmotd-372
type rplMotd struct {
	client string
	text   string
}

func (r rplMotd) format() string {
	return fmt.Sprintf(
		"372 %s :%s",
		r.client, r.text,
	)
}

// 375 RPL_MOTDSTART
// https://modern.ircdocs.horse/#rplmotdstart-375
type rplMotdStart struct {
	client string
	server string
	text   string
}

func (r rplMotdStart) format() string {
	return fmt.Sprintf(
		"375 %s :- %s %s",
		r.client, r.server, r.text,
	)
}

// 376 RPL_ENDOFMOTD
// https://modern.ircdocs.horse/#rplendofmotd-376
type rplEndOfMotd struct {
	client string
}

func (r rplEndOfMotd) format() string {
	return fmt.Sprintf(
		"376 %s :End of /MOTD command.",
		r.client,
	)
}

// 401 ERR_NOSUCHNICK
// https://modern.ircdocs.horse/#errnosuchnick-401
type errNoSuchNick struct {
	client string
	nick   string
}

func (r errNoSuchNick) format() string {
	return fmt.Sprintf(
		"401 %s %s :No such nickname.",
		r.client, r.nick,
	)
}

// 403 ERR_NOSUCHCHANNEL
// https://modern.ircdocs.horse/#errnosuchchannel-403
type errNoSuchChannel struct {
	client  string
	channel string
}

func (r errNoSuchChannel) format() string {
	return fmt.Sprintf(
		"403 %s %s :No such channel.",
		r.client, r.channel,
	)
}

// 432 ERR_ERRONEUSNICKNAME
// https://modern.ircdocs.horse/#errerroneusnickname-432
type errErroneusNickname struct {
	client string
	nick   string
}

func (r errErroneusNickname) format() string {
	return fmt.Sprintf(
		"432 %s %s :Erroneus nickname.",
		r.client, r.nick,
	)
}

// 433 ERR_NICKNAMEINUSE
// https://modern.ircdocs.horse/#errnicknameinuse-433
type errNicknameInUse struct {
	client string
	nick   string
}

func (r errNicknameInUse) format() string {
	return fmt.Sprintf(
		"433 %s %s :Nickname is already in use.",
		r.client, r.nick,
	)
}

// 442 ERR_NOTONCHANNEL
// https://modern.ircdocs.horse/#errnotonchannel-442
type errNotOnChannel struct {
	client  string
	channel string
}

func (r errNotOnChannel) format() string {
	return fmt.Sprintf(
		"442 %s %s :You are not on that channel.",
		r.client, r.channel,
	)
}

// 451 ERR_NOTREGISTERED
// https://modern.ircdocs.horse/#errnotregistered-451
type errNotRegistered struct {
	client string
}

func (r errNotRegistered) format() string {
	return fmt.Sprintf(
		"451 %s :You have not registered.",
		r.client,
	)
}

// 461 ERR_NEEDMOREPARAMS
// https://modern.ircdocs.horse/#errneedmoreparams-461
type errNeedMoreParams struct {
	client  string
	command string
}

func (r errNeedMoreParams) format() string {
	return fmt.Sprintf(
		"461 %s %s :Not enough parameters.",
		r.client, r.command,
	)
}

// 462 ERR_ALREADYREGISTERED
// https://modern.ircdocs.horse/#erralreadyregistered-462
type errAlreadyRegistered struct {
	client string
}

func (r errAlreadyRegistered) format() string {
	return fmt.Sprintf(
		"462 %s :You may not reregister.",
		r.client,
	)
}

// 475 ERR_BADCHANNELKEY
// https://modern.ircdocs.horse/#errbadchannelkey-475
type errBadChannelKey struct {
	client  string
	channel string
}

func (r errBadChannelKey) format() string {
	return fmt.Sprintf(
		"475 %s %s :Bad channel key (+k).",
		r.client, r.channel,
	)
}

// 502 ERR_USERSDONTMATCH
// https://modern.ircdocs.horse/#errusersdontmatch-502
type errUsersDontMatch struct {
	client string
}

func (r errUsersDontMatch) format() string {
	return fmt.Sprintf(
		"502 %s: Can't change mode for other users.",
		r.client,
	)
}
