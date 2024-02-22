package ircd

// client modes
type clientMode uint8

var clientModeMap = map[rune]clientMode{
	'i': modeClientInvisible,
	'o': modeClientOperator,
	'r': modeClientRegistered,
	'w': modeClientWallops,
}

const (
	modeClientInvisible = clientMode(1) << iota
	modeClientOperator
	modeClientRegistered
	modeClientWallops
)

// channel modes
type channelMode uint8

var channelModeMap = map[rune]channelMode{
	'i': modeChannelInvite,
	'k': modeChannelKey,
	'm': modeChannelModerated,
	's': modeChannelSecret,
	'p': modeChannelProtected,
}

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

func parseModestring[T ~uint8](modestring string, m map[rune]T) (add []T, del []T) {
	q := true

	for _, c := range modestring {
		switch c {
		case '+':
			q = true
			continue
		case '-':
			q = false
			continue
		default:
			m, ok := m[c]
			if !ok {
				continue
			}
			if q {
				add = append(add, m)
			}
			if !q {
				del = append(del, m)
			}
		}
	}

	return add, del
}
