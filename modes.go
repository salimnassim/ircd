package ircd

// client modes
type clientMode uint16

var clientModeMap = map[rune]clientMode{
	'i': modeClientInvisible,
	'o': modeClientOperator,
	'r': modeClientRegistered,
	'w': modeClientWallops,
	't': modeClientVhost,
}

const (
	modeClientInvisible = clientMode(1) << iota
	modeClientOperator
	modeClientRegistered
	modeClientWallops
	modeClientVhost
)

// channel modes
type channelMode uint16

var channelModeMap = map[rune]channelMode{
	'i': modeChannelInvite,
	'k': modeChannelKey,
	'm': modeChannelModerated,
	's': modeChannelSecret,
	'p': modeChannelPrivate,
	'C': modeChannelNoCTCP,
	'r': modeChannelRegistered,
	'O': modeChannelOpsOnly,
	'R': modeChannelRegisteredOnly,
	'n': modeChannelNoExternal,
}

const (
	modeChannelInvite = channelMode(1) << iota
	modeChannelKey
	modeChannelModerated
	modeChannelSecret
	modeChannelPrivate
	modeChannelNoCTCP
	modeChannelRegistered
	modeChannelOpsOnly
	modeChannelRegisteredOnly
	modeChannelNoExternal
)

type membershipMode uint16

var membershipModeMap = map[rune]membershipMode{
	'v': modeVoice,
	'h': modeHalfOperator,
	'o': modeOperator,
	'a': modeAdmin,
	'q': modeOwner,
}

// channel membership modes
const (
	modeVoice = membershipMode(1) << iota
	modeHalfOperator
	modeOperator
	modeAdmin
	modeOwner
)

func parseModestring[T ~uint16](modestring string, m map[rune]T) (add []T, del []T) {
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

// Finds differences between old and new mode bitmasks.
// Add represents modes that been added from the original list..
// Del represents modes that been removed from the original list.
func diffModes[T ~uint16](old T, new T, m map[rune]T) (add []T, del []T) {
	d := old ^ new

	for _, b := range m {
		if d&b != 0 {
			if new&b != 0 {
				add = append(add, b)
			} else {
				del = append(del, b)
			}
		}
	}

	return add, del
}