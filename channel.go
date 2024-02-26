package ircd

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
	"time"
)

type channeler interface {
	name() string
	owner() clientID

	clients() channelClientStorer
	count() int

	password() string
	setPassword(password string)

	secret() bool
	setSecret(secret bool)

	topic() *topic
	setTopic(text string, author string)

	// Does prefix match any of the ban masks?
	banned(c clienter) bool
	// Add ban mask.
	addBan(mask banMask) error
	// Remove ban mask.
	removeBan(mask banMask) error

	addClient(c clienter, password string) error
	removeClient(c clienter)

	names() []string

	broadcastRPL(rpl rpl, sourceID clientID, skip bool)
	broadcastCommand(cmd command, sourceID clientID, skip bool)

	modestring() string
	addMode(mode channelMode)
	removeMode(mode channelMode)
	hasMode(mode channelMode) bool
}

type banMask string

type channel struct {
	mu *sync.RWMutex
	// Channel name.
	n string
	// Channel topic.
	t *topic
	// Channel clients.
	cs    channelClientStorer
	modes channelMode
	bans  map[banMask]bool
	// Channel owner.
	o clientID
	// Channel password.
	p string
	// Is channel secret?
	s bool
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
		cs:    newChannelClientStore(),
		modes: 0,
		o:     owner,
		p:     "",
		s:     false,
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

func (ch *channel) password() string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.p
}

func (ch *channel) setPassword(password string) {
	ch.mu.Lock()
	ch.p = password
	ch.mu.Unlock()
}

func (ch *channel) secret() bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.s
}

func (ch *channel) setSecret(secret bool) {
	ch.mu.Lock()
	ch.s = secret
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

	for mask, _ := range ch.bans {
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

// Adds client to channel. If password does not match, an error is returned.
func (ch *channel) addClient(c clienter, password string) error {
	if password != "" && ch.p != password {
		return errorBadChannelKey
	}

	ch.cs.add(c)

	return nil
}

// Remove client from channel.
func (ch *channel) removeClient(c clienter) {
	ch.cs.delete(c.id())
}

// Returns channel users delimited by a space for RPL_NAMREPLY.
func (ch *channel) names() []string {
	var names []string

	for _, c := range ch.cs.all() {
		if ch.o == c.id() {
			names = append(names, fmt.Sprintf("@%s", c.nickname()))
		} else {
			names = append(names, c.nickname())
		}
	}

	return names
}

// Send RPL to all clients on the channel.
// If skip is true, the client in source will not receive the message.
func (ch *channel) broadcastRPL(rpl rpl, sourceID clientID, skip bool) {
	for _, c := range ch.cs.all() {
		if c.id() == sourceID && skip {
			continue
		}
		c.send(rpl.format())
	}
}

// Send command to all clients on the channel.
// If skip is true, the client in source will not receive the message.
func (ch *channel) broadcastCommand(cmd command, sourceID clientID, skip bool) {
	for _, c := range ch.cs.all() {
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
