package ircd

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
	"time"
)

type channeler interface {
	// Channel name.
	name() string
	// Channel owner.
	owner() clientID

	// Access client members.
	clients() channelClientStorer
	// Number of channel members.
	count() int

	// Channel topic.
	topic() *topic
	// Set channel topic.
	setTopic(text string, author string)

	// Does prefix match any of the ban masks?
	banned(c clienter) bool
	// Add ban mask.
	addBan(mask banMask) error
	// Remove ban mask.
	removeBan(mask banMask) error

	// Channel members in NAMES format including highest prefix.
	names() []string

	// Broadcast RPL to channel members.
	//
	// If skip is set to true, the source client will not receive the RPL message.
	broadcastRPL(rpl rpl, sourceID clientID, skip bool)
	// Broadcast command to channel members.
	//
	// If skip is set to true, the source client will not receive the command message.
	broadcastCommand(cmd command, sourceID clientID, skip bool)

	// Channel modestring.
	modestring() string

	// Channel modestring as a bitmask.
	mode() (mode channelMode)
	// Add mode to channel.
	addMode(mode channelMode)
	// Remove mode from chanel.
	removeMode(mode channelMode)
	// Does channel have mode?
	hasMode(mode channelMode) bool

	// Has client been invited to the channel?
	isInvited(c clienter) bool
	// Add invite to channel
	addInvite(clientID clientID)
	// Remove invite from channel
	removeInvite(clientID clientID)

	// Get channel key (password).
	key() string
	// Set channel key (password).
	setKey(key string)
}

type banMask string

type channel struct {
	mu *sync.RWMutex
	// Channel name.
	n string
	// Channel topic.
	t *topic
	// Channel clients.
	cs      channelClientStorer
	modes   channelMode
	bans    map[banMask]bool
	invites map[clientID]bool
	// Channel owner.
	o clientID
	// Channel password.
	k string
}

type topic struct {
	text      string
	timestamp int
	author    string
}

func newChannel(channelName string, owner clientID) *channel {
	channel := &channel{
		mu: &sync.RWMutex{},
		n:  channelName,
		t: &topic{
			text:      "",
			timestamp: 0,
			author:    "",
		},
		cs:      newChannelClientStore(),
		modes:   0,
		bans:    make(map[banMask]bool),
		invites: make(map[clientID]bool),
		o:       owner,
		k:       "",
	}

	return channel
}

func (ch *channel) name() string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.n
}

func (ch *channel) clients() channelClientStorer {
	return ch.cs
}

func (ch *channel) count() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.cs.count()
}

func (ch *channel) key() string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.k
}

func (ch *channel) setKey(key string) {
	ch.mu.Lock()
	ch.k = key
	ch.mu.Unlock()
}

func (ch *channel) owner() clientID {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.o
}

// Sets channel topic.
func (ch *channel) setTopic(text string, author string) {
	ch.mu.Lock()
	ch.t.text = text
	ch.t.timestamp = int(time.Now().Unix())
	ch.t.author = author
	ch.mu.Unlock()
}

func (ch *channel) banned(c clienter) bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	for mask := range ch.bans {
		if matchMask([]byte(mask), c.hostname()) {
			return true
		}
	}
	return false
}

func (ch *channel) addBan(mask banMask) error {
	ch.mu.RLock()
	_, ok := ch.bans[mask]
	ch.mu.RUnlock()
	if ok {
		return errorBanMaskAlreadyExists
	}
	ch.mu.Lock()
	ch.bans[mask] = true
	ch.mu.Unlock()
	return nil
}

func (ch *channel) removeBan(mask banMask) error {
	ch.mu.RLock()
	_, ok := ch.bans[mask]
	ch.mu.RUnlock()
	if !ok {
		return errorBanMaskDoesNotExist
	}
	ch.mu.Lock()
	delete(ch.bans, mask)
	ch.mu.Unlock()
	return nil
}

// Returns current topic.
func (ch *channel) topic() *topic {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.t
}

// Returns channel users delimited by a space for RPL_NAMREPLY.
func (ch *channel) names() []string {
	var names []string

	for _, c := range ch.cs.all() {
		names = append(names, c.nickname())
	}

	return names
}

// Send RPL to all clients on the channel.
// If skip is true, the client in source will not receive the message.
func (ch *channel) broadcastRPL(rpl rpl, sourceID clientID, skip bool) {
	for _, c := range ch.cs.all() {
		// Do not broadcast to clients that are quitting.
		if c.quitReason() != "" {
			continue
		}
		if c.id() == sourceID && skip {
			continue
		}
		c.send(rpl.rpl())
	}
}

// Send command to all clients on the channel.
// If skip is true, the client in source will not receive the message.
func (ch *channel) broadcastCommand(cmd command, sourceID clientID, skip bool) {
	for _, c := range ch.cs.all() {
		// Do not broadcast to clients that are quitting.
		if c.quitReason() != "" {
			continue
		}
		if c.id() == sourceID && skip {
			continue
		}
		c.send(cmd.command())
	}
}

func (ch *channel) modestring() string {
	modes := []rune{}
	for m, r := range channelModeMap {
		if ch.hasMode(r) {
			modes = append(modes, m)
		}
	}
	slices.SortFunc[[]rune, rune](modes, func(a rune, b rune) int {
		return cmp.Compare(a, b)
	})
	return fmt.Sprintf("+%s", string(modes))
}

func (ch *channel) mode() (mode channelMode) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.modes
}

func (ch *channel) addMode(mode channelMode) {
	if ch.hasMode(mode) {
		return
	}

	ch.mu.Lock()
	ch.modes |= mode
	ch.mu.Unlock()
}

func (ch *channel) removeMode(mode channelMode) {
	if !ch.hasMode(mode) {
		return
	}

	ch.mu.Lock()
	ch.modes &= ^mode
	ch.mu.Unlock()
}

func (ch *channel) hasMode(mode channelMode) bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.modes&mode != 0
}

func (ch *channel) isInvited(c clienter) bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	// is client in the invite map?
	iv, ok := ch.invites[c.id()]
	if !ok {
		return false
	}
	// invite has been already used
	if !iv {
		return false
	}
	return true
}

func (ch *channel) addInvite(clientID clientID) {
	ch.mu.Lock()
	ch.invites[clientID] = true
	ch.mu.Unlock()
}

func (ch *channel) removeInvite(clientID clientID) {
	ch.mu.Lock()
	delete(ch.invites, clientID)
	ch.mu.Unlock()
}
