package ircd

// client modes
type clientMode uint8

const (
	// +i
	modeClientInvisible = clientMode(1) << iota
	// +o
	modeClientOperator
	// +r
	modeClientRegistered
	// +w
	modeClientWallops
)

// channel modes
type channelMode uint8

const (
	// +i
	modeChannelInvite = channelMode(1) << iota
	// +k
	modeChannelKey
	// +m
	modeChannelModerated
	// +s
	modeChannelSecret
	// +t
	modeChannelProtected
)

type membershipMode uint8

// channel membership modes
const (
	// +v
	modeVoice = membershipMode(1) << iota
	// +h
	modeHalfOperator
	// +o
	modeOperator
	// +q
	modeFounder
	// +a
	modeProtected
)
