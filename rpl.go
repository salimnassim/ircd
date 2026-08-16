package ircd

import (
	"fmt"
	"strings"
)

type rpl interface {
	rpl() string
}

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

type rplLuserClient struct {
	client string

	users int

	invisible int

	servers int
}

func (r rplLuserClient) rpl() string {
	return fmt.Sprintf(
		"251 %s :There are %d users (%d invisible) on %d servers",
		r.client, r.users, r.invisible, r.servers,
	)
}

type rplLuserOp struct {
	client string

	ops int
}

func (r rplLuserOp) rpl() string {
	return fmt.Sprintf(
		"252 %s %d :operator(s) online",
		r.client, r.ops,
	)
}

type rplLuserChannels struct {
	client string

	channels int
}

func (r rplLuserChannels) rpl() string {
	return fmt.Sprintf(
		"254 %s %d :channels formed.",
		r.client, r.channels,
	)
}

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

type rplUnAway struct {
	client string
}

func (r rplUnAway) rpl() string {
	return fmt.Sprintf(
		"305 %s :You are no longer marked as being away.",
		r.client,
	)
}

type rplNowAway struct {
	client string
}

func (r rplNowAway) rpl() string {
	return fmt.Sprintf(
		"306 %s :You have been marked as being away.",
		r.client,
	)
}

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

type rplListStart struct {
	client string
}

func (r rplListStart) rpl() string {
	return fmt.Sprintf(
		"321 %s Channel :Users Name",
		r.client,
	)
}

type rplList struct {
	client  string
	channel string

	count int
	topic string
}

func (r rplList) rpl() string {
	return fmt.Sprintf(
		"322 %s %s %d :%s",
		r.client, r.channel, r.count, r.topic,
	)
}

type rplListEnd struct {
	client string
}

func (r rplListEnd) rpl() string {
	return fmt.Sprintf(
		"323 %s :End of /LIST",
		r.client,
	)
}

type rplChannelModeIs struct {
	client     string
	channel    string
	modestring string
	modeargs   string
}

func (r rplChannelModeIs) rpl() string {
	if r.modeargs == "" {
		return fmt.Sprintf(
			"324 %s %s %s",
			r.client, r.channel, r.modestring,
		)
	}
	return fmt.Sprintf(
		"324 %s %s %s %s",
		r.client, r.channel, r.modestring, r.modeargs,
	)
}

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

type rplNamReply struct {
	client string

	symbol string

	channel string

	nicks []string
}

func (r rplNamReply) rpl() string {
	nicks := strings.Join(r.nicks, " ")
	return fmt.Sprintf(
		"353 %s %s %s :%s",
		r.client, r.symbol, r.channel, nicks,
	)
}

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

type rplBanList struct {
	client  string
	channel string
	mask    string
}

func (r rplBanList) rpl() string {
	return fmt.Sprintf(
		"367 %s %s %s",
		r.client, r.channel, r.mask,
	)
}

type rplEndOfBanList struct {
	client  string
	channel string
}

func (r rplEndOfBanList) rpl() string {
	return fmt.Sprintf(
		"368 %s %s :End of channel ban list.",
		r.client, r.channel,
	)
}

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

type rplEndOfMotd struct {
	client string
}

func (r rplEndOfMotd) rpl() string {
	return fmt.Sprintf(
		"376 %s :End of /MOTD command.",
		r.client,
	)
}

type rplYoureOper struct {
	client string
}

func (r rplYoureOper) rpl() string {
	return fmt.Sprintf(
		"381 %s :You are now an IRC operator.",
		r.client,
	)
}

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

type errNoNicknameGiven struct {
	client string
}

func (r errNoNicknameGiven) rpl() string {
	return fmt.Sprintf(
		"431 %s :No nickname given.",
		r.client,
	)
}

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

type errNotRegistered struct {
	client string
}

func (r errNotRegistered) rpl() string {
	return fmt.Sprintf(
		"451 %s :You have not registered.",
		r.client,
	)
}

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

type errAlreadyRegistered struct {
	client string
}

func (r errAlreadyRegistered) rpl() string {
	return fmt.Sprintf(
		"462 %s :You may not reregister.",
		r.client,
	)
}

type errPasswdMismatch struct {
	client string
}

func (r errPasswdMismatch) rpl() string {
	return fmt.Sprintf(
		"464 %s :Password incorrect.",
		r.client,
	)
}

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

type errBannedFromChan struct {
	client  string
	channel string
}

func (r errBannedFromChan) rpl() string {
	return fmt.Sprintf(
		"474 %s %s :Cannot join channel (+b)",
		r.client, r.channel,
	)
}

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

type errNoPrivileges struct {
	client string
}

func (r errNoPrivileges) rpl() string {
	return fmt.Sprintf(
		"481 %s :Permission Denied - You're not an IRC operator.",
		r.client,
	)
}

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

type errUsersDontMatch struct {
	client string
}

func (r errUsersDontMatch) rpl() string {
	return fmt.Sprintf(
		"502 %s :Can't change mode for other users.",
		r.client,
	)
}

type errUnknownCommand struct {
	client  string
	command string
}

func (r errUnknownCommand) rpl() string {
	return fmt.Sprintf(
		"421 %s %s :Unknown command.",
		r.client, r.command,
	)
}

type errTooManyChannels struct {
	client  string
	channel string
}

func (r errTooManyChannels) rpl() string {
	return fmt.Sprintf(
		"405 %s %s :You have joined too many channels.",
		r.client, r.channel,
	)
}
