package ircd

type channelEvent interface {
	isChannelEvent()
}

type evJoin struct {
	who *sessionHandle
	key string
}

func (evJoin) isChannelEvent() {}

type evPart struct {
	who    *sessionHandle
	reason string
}

func (evPart) isChannelEvent() {}

type evKick struct {
	by     *sessionHandle
	target *sessionHandle
	reason string
}

func (evKick) isChannelEvent() {}

type evSetTopic struct {
	by   *sessionHandle
	text string
}

func (evSetTopic) isChannelEvent() {}

type evSetMode struct {
	by         *sessionHandle
	modestring string
	args       string
}

func (evSetMode) isChannelEvent() {}

type broadcastKind int

const (
	broadcastPrivmsg broadcastKind = iota
	broadcastNotice
)

type evBroadcast struct {
	by   *sessionHandle
	text string
	kind broadcastKind
}

func (evBroadcast) isChannelEvent() {}

type evInvite struct {
	by     *sessionHandle
	target *sessionHandle
}

func (evInvite) isChannelEvent() {}

type evQuit struct {
	who    *sessionHandle
	reason string
}

func (evQuit) isChannelEvent() {}

type evMemberRenamed struct {
	id        clientID
	oldPrefix string
	newNick   string
}

func (evMemberRenamed) isChannelEvent() {}
