package ircd

// client modes
type clientMode uint8

const (
	modeClientInvisible = clientMode(1) << iota
	modeClientOperator
	modeClientRegistered
	modeClientWallops
)

// channel modes
type channelMode uint8

const (
	modeChannelInvite = channelMode(1) << iota
	modeChannelKey
	modeChannelModerated
	modeChannelSecret
	modeChannelProtected
)

type membershipMode uint8

// channel membership modes
const (
	modeVoice = membershipMode(1) << iota
	modeHalfOperator
	modeOperator
	modeFounder
	modeProtected
)
