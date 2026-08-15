package ircd

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
	modeMemberVoice = channelMembershipMode(1) << iota

	modeMemberHalfOperator

	modeMemberOperator

	modeMemberAdmin

	modeMemberOwner
)

// membershipPrefixOrder lists membership modes from highest
// to lowest privilege.
var membershipPrefixOrder = []struct {
	mode   channelMembershipMode
	prefix rune
}{
	{modeMemberOwner, '~'},
	{modeMemberAdmin, '&'},
	{modeMemberOperator, '@'},
	{modeMemberHalfOperator, '%'},
	{modeMemberVoice, '+'},
}

// membershipPrefix returns the symbol for the highest-privilege mode set in
// modes, or 0 if none apply.
func membershipPrefix(modes channelMembershipMode) rune {
	for _, p := range membershipPrefixOrder {
		if modes&p.mode != 0 {
			return p.prefix
		}
	}
	return 0
}

func runeByMode[T ~uint16](t T, m map[rune]T) rune {
	for r, mode := range m {
		if t == mode {
			return r
		}
	}
	return '?'
}

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
