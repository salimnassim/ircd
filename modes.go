package ircd

// client modes
type clientMode uint16

var clientModeMap = map[rune]clientMode{
	'i': modeClientInvisible,
	'o': modeClientOperator,
	'r': modeClientRegistered,
	'w': modeClientWallops,
	't': modeClientVhost,
	'z': modeClientTLS,
}

const (
	modeClientInvisible = clientMode(1) << iota
	modeClientOperator
	modeClientRegistered
	modeClientWallops
	modeClientVhost
	modeClientTLS
)

// channel modes
type channelMode uint16

var channelModeMap = map[rune]channelMode{
	'i': modeChannelInviteOnly,
	'k': modeChannelKey,
	'm': modeChannelModerated,
	's': modeChannelSecret,
	'p': modeChannelPrivate,
	'C': modeChannelNoCTCP,
	'r': modeChannelRegistered,
	'O': modeChannelOpsOnly,
	'R': modeChannelRegisteredOnly,
	'n': modeChannelNoExternal,
	'z': modeChannelTLSOnly,
	't': modeChannelRestrictTopic,
}

const (
	modeChannelInviteOnly = channelMode(1) << iota
	modeChannelKey
	modeChannelModerated
	modeChannelSecret
	modeChannelPrivate
	modeChannelNoCTCP
	modeChannelRegistered
	modeChannelOpsOnly
	modeChannelRegisteredOnly
	modeChannelNoExternal
	modeChannelTLSOnly
	modeChannelRestrictTopic
)

type channelMembershipMode uint16

var channelMembershipModeMap = map[rune]channelMembershipMode{
	'v': modeMemberVoice,
	'h': modeMemberHalfOperator,
	'o': modeMemberOperator,
	'a': modeMemberAdmin,
	'q': modeMemberOwner,
}

const (
	// Channel voiced.
	modeMemberVoice = channelMembershipMode(1) << iota
	// Channel half-operator.
	modeMemberHalfOperator
	// Channel operator.
	modeMemberOperator
	// Channel admin.
	modeMemberAdmin
	// Channel owner.
	modeMemberOwner
)

// Get matching rune from map by T.
func runeByMode[T ~uint16](t T, m map[rune]T) rune {
	for r, mode := range m {
		if t == mode {
			return r
		}
	}
	return '?'
}

// Parse modestring into add/del slices of type T.
//
// Example: +v-o for channelMembershipModeMap becomes [modeMemberVoice], [modeMemberOperator]
func parseModestring[T ~uint16](modestring string, m map[rune]T) (add []T, del []T) {
	q := true
	add = []T{}
	del = []T{}

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
//
// Add represents modes that been added from the original list..
//
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
