package ircd

import (
	"fmt"
	"strings"
)

type rpl interface {
	rpl() string
}

// 001 RPL_WELCOME
//
// https://modern.ircdocs.horse/#rplwelcome-001
type rplWelcome struct {
	client   string
	network  string
	hostname string
}

func (r rplWelcome) rpl() string {
	return fmt.Sprintf(
		"001 %s :Welcome to the %s Network, %s",
		r.client, r.network, r.hostname,
	)
}

// 002 RPL_YOURHOST
//
// https://modern.ircdocs.horse/#rplyourhost-002
type rplYourHost struct {
	client     string
	serverName string
	version    string
}

func (r rplYourHost) rpl() string {
	return fmt.Sprintf(
		"002 %s :Your host is %s, running version %s",
		r.client, r.serverName, r.version,
	)
}

// 005 RPL_ISUPPORT
//
// https://modern.ircdocs.horse/#rplisupport-005
type rplISupport struct {
	client string
	tokens string
}

func (r rplISupport) rpl() string {
	return fmt.Sprintf(
		"005 %s %s :are supported by this server.",
		r.client, r.tokens,
	)
}

// 221 RPL_UMODEIS
//
// https://modern.ircdocs.horse/#rplumodeis-221
type rplUModeIs struct {
	client     string
	modestring string
}

func (r rplUModeIs) rpl() string {
	return fmt.Sprintf(
		"221 %s %s",
		r.client, r.modestring,
	)
}

// 251 RPL_LUSERCLIENT.
//
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

func (r rplLuserClient) rpl() string {
	return fmt.Sprintf(
		"251 %s :There are %d users (%d invisible) on %d servers",
		r.client, r.users, r.invisible, r.servers,
	)
}

// 252 RPL_LUSEROP
//
// https://modern.ircdocs.horse/#rplluserop-252
type rplLuserOp struct {
	client string
	// Number of operators.
	ops int
}

func (r rplLuserOp) rpl() string {
	return fmt.Sprintf(
		"252 %s %d :operator(s) online",
		r.client, r.ops,
	)
}

// 254 RPL_LUSERCHANNELS
//
// https://modern.ircdocs.horse/#rplluserchannels-254
type rplLuserChannels struct {
	client string
	// Number of channels.
	channels int
}

func (r rplLuserChannels) rpl() string {
	return fmt.Sprintf(
		"254 %s %d :channels formed.",
		r.client, r.channels,
	)
}

// 301 RPL_AWAY
//
// https://modern.ircdocs.horse/#rplaway-301
type rplAway struct {
	client  string
	nick    string
	message string
}

func (r rplAway) rpl() string {
	return fmt.Sprintf(
		"301 %s %s :%s",
		r.client, r.nick, r.message,
	)
}

// 305 RPL_UNAWAY
//
// https://modern.ircdocs.horse/#rplunaway-305
type rplUnAway struct {
	client string
}

func (r rplUnAway) rpl() string {
	return fmt.Sprintf(
		"305 %s :You are no longer marked as being away.",
		r.client,
	)
}

// 306 RPL_NOWAWAY
//
// https://modern.ircdocs.horse/#rplnowaway-306
type rplNowAway struct {
	client string
}

func (r rplNowAway) rpl() string {
	return fmt.Sprintf(
		"306 %s :You have been marked as being away.",
		r.client,
	)
}

// 311 RPL_WHOISUSER
//
// https://modern.ircdocs.horse/#rplwhoisuser-311
type rplWhoisUser struct {
	client   string
	nick     string
	username string
	host     string
	realname string
}

func (r rplWhoisUser) rpl() string {
	return fmt.Sprintf(
		"311 %s %s %s %s * :%s",
		r.client, r.nick, r.username, r.host, r.realname,
	)
}

// 315 RPL_ENDOFWHO
//
// https://modern.ircdocs.horse/#rplendofwho-315
type rplEndOfWho struct {
	client string
	mask   string
}

func (r rplEndOfWho) rpl() string {
	return fmt.Sprintf(
		"315 %s %s :End of WHO list.",
		r.client, r.mask,
	)
}

// 319 RPL_WHOISCHANNELS
//
// https://modern.ircdocs.horse/#rplwhoischannels-319
type rplWhoisChannels struct {
	client   string
	nick     string
	channels []string
}

func (r rplWhoisChannels) rpl() string {
	channels := strings.Join(r.channels, " ")
	return fmt.Sprintf(
		"319 %s %s :%s",
		r.client, r.nick, channels,
	)
}

// 320 RPL_WHOISSPECIAL
//
// https://modern.ircdocs.horse/#rplwhoisspecial-320
type rplWhoisSpecial struct {
	client string
	nick   string
	text   string
}

func (r rplWhoisSpecial) rpl() string {
	return fmt.Sprintf(
		"320 %s %s :%s",
		r.client, r.nick, r.text,
	)
}

// 321 RPL_LISTSTART
//
// https://modern.ircdocs.horse/#rplliststart-321
type rplListStart struct {
	client string
}

func (r rplListStart) rpl() string {
	return fmt.Sprintf(
		"321 %s Channel :Users Name",
		r.client,
	)
}

// 322 RPL_LIST
//
// https://modern.ircdocs.horse/#rpllist-322
type rplList struct {
	client  string
	channel string
	// Number of clients on server.
	count int
	topic string
}

func (r rplList) rpl() string {
	return fmt.Sprintf(
		"322 %s %s %d :%s",
		r.client, r.channel, r.count, r.topic,
	)
}

// 323 RPL_LISTEND
//
// https://modern.ircdocs.horse/#rpllistend-323
type rplListEnd struct {
	client string
}

func (r rplListEnd) rpl() string {
	return fmt.Sprintf(
		"323 %s :End of /LIST",
		r.client,
	)
}

// 324 RPL_CHANNELMODEIS
//
// https://modern.ircdocs.horse/#rplchannelmodeis-324
type rplChannelModeIs struct {
	client     string
	channel    string
	modestring string
	modeargs   string
}

func (r rplChannelModeIs) rpl() string {
	return fmt.Sprintf(
		"324 %s %s %s %s",
		r.client, r.channel, r.modestring, r.modeargs,
	)
}

// 331 RPL_NOTOPIC
//
// https://modern.ircdocs.horse/#rplnotopic-331
type rplNoTopic struct {
	client  string
	channel string
}

func (r rplNoTopic) rpl() string {
	return fmt.Sprintf(
		"331 %s %s :No topic is set.",
		r.client, r.channel,
	)
}

// 331 RPL_TOPIC
//
// https://modern.ircdocs.horse/#rpltopic-332
type rplTopic struct {
	client  string
	channel string
	topic   string
}

func (r rplTopic) rpl() string {
	return fmt.Sprintf(
		"332 %s %s :%s",
		r.client, r.channel, r.topic,
	)
}

// 333 RPL_TOPICWHOTIME
//
// https://modern.ircdocs.horse/#rpltopicwhotime-333
type rplTopicWhoTime struct {
	client  string
	channel string
	nick    string
	setat   int
}

func (r rplTopicWhoTime) rpl() string {
	return fmt.Sprintf(
		"333 %s %s %s %d",
		r.client, r.channel, r.nick, r.setat,
	)
}

// 341 RPL_INVITING
//
// https://modern.ircdocs.horse/#rplinviting-341
type rplInviting struct {
	client  string
	nick    string
	channel string
}

func (r rplInviting) rpl() string {
	return fmt.Sprintf(
		"341 %s %s %s",
		r.client, r.nick, r.channel,
	)
}

// 351 RPL_VERSION
//
// https://modern.ircdocs.horse/#rplversion-351
type rplVersion struct {
	client   string
	version  string
	server   string
	comments string
}

func (r rplVersion) rpl() string {
	return fmt.Sprintf(
		"351 %s %s %s :%s",
		r.client, r.version, r.server, r.comments,
	)
}

// 352 RPL_WHOREPLY
//
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

func (r rplWhoReply) rpl() string {
	return fmt.Sprintf(
		"352 %s %s %s %s %s %s %s :%d %s",
		r.client, r.channel, r.username, r.host, r.server, r.nick, r.flags, r.hopcount, r.realname,
	)
}

// 353 RPL_NAMREPLY.
//
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

func (r rplNamReply) rpl() string {
	nicks := strings.Join(r.nicks, " ")
	return fmt.Sprintf(
		"353 %s %s %s :%s",
		r.client, r.symbol, r.channel, nicks,
	)
}

// 366 RPL_ENDOFNAMES
//
// https://modern.ircdocs.horse/#rplendofnames-366
type rplEndOfNames struct {
	client  string
	channel string
}

func (r rplEndOfNames) rpl() string {
	return fmt.Sprintf(
		"366 %s %s :End of /NAMES list.",
		r.client, r.channel,
	)
}

// 372 RPL_MOTD
//
// https://modern.ircdocs.horse/#rplmotd-372
type rplMotd struct {
	client string
	text   string
}

func (r rplMotd) rpl() string {
	return fmt.Sprintf(
		"372 %s :%s",
		r.client, r.text,
	)
}

// 375 RPL_MOTDSTART
//
// https://modern.ircdocs.horse/#rplmotdstart-375
type rplMotdStart struct {
	client string
	server string
	text   string
}

func (r rplMotdStart) rpl() string {
	return fmt.Sprintf(
		"375 %s :- %s %s",
		r.client, r.server, r.text,
	)
}

// 376 RPL_ENDOFMOTD
//
// https://modern.ircdocs.horse/#rplendofmotd-376
type rplEndOfMotd struct {
	client string
}

func (r rplEndOfMotd) rpl() string {
	return fmt.Sprintf(
		"376 %s :End of /MOTD command.",
		r.client,
	)
}

// 381 RPL_YOUREOPER
//
// https://modern.ircdocs.horse/#rplyoureoper-381
type rplYoureOper struct {
	client string
}

func (r rplYoureOper) rpl() string {
	return fmt.Sprintf(
		"381 %s :You are now an IRC operator.",
		r.client,
	)
}

// 401 ERR_NOSUCHNICK
//
// https://modern.ircdocs.horse/#errnosuchnick-401
type errNoSuchNick struct {
	client string
	nick   string
}

func (r errNoSuchNick) rpl() string {
	return fmt.Sprintf(
		"401 %s %s :No such nickname.",
		r.client, r.nick,
	)
}

// 402 ERR_NOSUCHSERVER
//
// https://modern.ircdocs.horse/#errnosuchserver-402
type errNoSuchServer struct {
	client string
	server string
}

func (r errNoSuchServer) rpl() string {
	return fmt.Sprintf(
		"402 %s %s :No such server or user.",
		r.client, r.server,
	)
}

// 403 ERR_NOSUCHCHANNEL
//
// https://modern.ircdocs.horse/#errnosuchchannel-403
type errNoSuchChannel struct {
	client  string
	channel string
}

func (r errNoSuchChannel) rpl() string {
	return fmt.Sprintf(
		"403 %s %s :No such channel.",
		r.client, r.channel,
	)
}

// 404 ERR_CANNOTSENDTOCHAN
//
// https://modern.ircdocs.horse/#errcannotsendtochan-404
type errCannotSendToChan struct {
	client  string
	channel string
	text    string
}

func (r errCannotSendToChan) rpl() string {
	return fmt.Sprintf(
		"404 %s %s :%s",
		r.client, r.channel, r.text,
	)
}

// 431 ERR_NONICKNAMEGIVEN
//
// https://modern.ircdocs.horse/#errnonicknamegiven-431
type errNoNicknameGiven struct {
	client string
}

func (r errNoNicknameGiven) rpl() string {
	return fmt.Sprintf(
		"431 %s :No nickname given.",
		r.client,
	)
}

// 432 ERR_ERRONEUSNICKNAME
//
// https://modern.ircdocs.horse/#errerroneusnickname-432
type errErroneusNickname struct {
	client string
	nick   string
}

func (r errErroneusNickname) rpl() string {
	return fmt.Sprintf(
		"432 %s %s :Erroneus nickname.",
		r.client, r.nick,
	)
}

// 433 ERR_NICKNAMEINUSE
//
// https://modern.ircdocs.horse/#errnicknameinuse-433
type errNicknameInUse struct {
	client string
	nick   string
}

func (r errNicknameInUse) rpl() string {
	return fmt.Sprintf(
		"433 %s %s :Nickname is already in use.",
		r.client, r.nick,
	)
}

// 441 ERR_USERNOTINCHANNEL
//
// https://modern.ircdocs.horse/#errusernotinchannel-441
type errUserNotInChannel struct {
	client  string
	nick    string
	channel string
}

func (r errUserNotInChannel) rpl() string {
	return fmt.Sprintf(
		"441 %s %s %s :They aren't on that channel.",
		r.client, r.nick, r.channel,
	)
}

// 442 ERR_NOTONCHANNEL
//
// https://modern.ircdocs.horse/#errnotonchannel-442
type errNotOnChannel struct {
	client  string
	channel string
}

func (r errNotOnChannel) rpl() string {
	return fmt.Sprintf(
		"442 %s %s :You are not on that channel.",
		r.client, r.channel,
	)
}

// 443 ERR_USERONCHANNEL
//
// https://modern.ircdocs.horse/#erruseronchannel-443
type errUserOnChannel struct {
	client  string
	nick    string
	channel string
}

func (r errUserOnChannel) rpl() string {
	return fmt.Sprintf(
		"443 %s %s %s :is already on channel.",
		r.client, r.nick, r.channel,
	)
}

// 451 ERR_NOTREGISTERED
//
// https://modern.ircdocs.horse/#errnotregistered-451
type errNotRegistered struct {
	client string
}

func (r errNotRegistered) rpl() string {
	return fmt.Sprintf(
		"451 %s :You have not registered.",
		r.client,
	)
}

// 461 ERR_NEEDMOREPARAMS
//
// https://modern.ircdocs.horse/#errneedmoreparams-461
type errNeedMoreParams struct {
	client  string
	command string
}

func (r errNeedMoreParams) rpl() string {
	return fmt.Sprintf(
		"461 %s %s :Not enough parameters.",
		r.client, r.command,
	)
}

// 462 ERR_ALREADYREGISTERED
//
// https://modern.ircdocs.horse/#erralreadyregistered-462
type errAlreadyRegistered struct {
	client string
}

func (r errAlreadyRegistered) rpl() string {
	return fmt.Sprintf(
		"462 %s :You may not reregister.",
		r.client,
	)
}

// 464 RR_PASSWDMISMATCH
//
// https://modern.ircdocs.horse/#errpasswdmismatch-464
type errPasswdMismatch struct {
	client string
}

func (r errPasswdMismatch) rpl() string {
	return fmt.Sprintf(
		"464 %s :Password incorrect.",
		r.client,
	)
}

// 473 ERR_INVITEONLYCHAN
//
// https://modern.ircdocs.horse/#errinviteonlychan-473
type errInviteOnlyChan struct {
	client  string
	channel string
}

func (r errInviteOnlyChan) rpl() string {
	return fmt.Sprintf(
		"473 %s %s :Cannot join channel (+i)",
		r.client, r.channel,
	)
}

// 474 ERR_BANNEDFROMCHAN
//
// https://modern.ircdocs.horse/#errbannedfromchan-474
type errBannedFromChan struct {
	client  string
	channel string
}

func (r errBannedFromChan) rpl() string {
	return fmt.Sprintf(
		"474 %s %s :Cannot join channel (+z)",
		r.client, r.channel,
	)
}

// 475 ERR_BADCHANNELKEY
//
// https://modern.ircdocs.horse/#errbadchannelkey-475
type errBadChannelKey struct {
	client  string
	channel string
}

func (r errBadChannelKey) rpl() string {
	return fmt.Sprintf(
		"475 %s %s :Bad channel key (+k).",
		r.client, r.channel,
	)
}

// 481 ERR_NOPRIVILEGES
//
// https://modern.ircdocs.horse/#errnoprivileges-481
type errNoPrivileges struct {
	client string
}

func (r errNoPrivileges) rpl() string {
	return fmt.Sprintf(
		"481 %s :Permission Denied - You're not an IRC operator.",
		r.client,
	)
}

// 482 ERR_CHANOPRIVSNEEDED
//
// https://modern.ircdocs.horse/#errchanoprivsneeded-482
type errChanoPrivsNeeded struct {
	client  string
	channel string
}

func (r errChanoPrivsNeeded) rpl() string {
	return fmt.Sprintf(
		"482 %s %s :You're not channel operator.",
		r.client, r.channel,
	)
}

// 502 ERR_USERSDONTMATCH
//
// https://modern.ircdocs.horse/#errusersdontmatch-502
type errUsersDontMatch struct {
	client string
}

func (r errUsersDontMatch) rpl() string {
	return fmt.Sprintf(
		"502 %s :Can't change mode for other users.",
		r.client,
	)
}

// 723 ERR_NOPRIVS
//
// https://modern.ircdocs.horse/#errnoprivs-723
// type errNoPrivs struct {
// 	client string
// 	priv   string
// }

// func (r errNoPrivs) format() string {
// 	return fmt.Sprintf(
// 		"723 %s %s :Insufficient oper privileges.",
// 		r.client, r.priv,
// 	)
// }
