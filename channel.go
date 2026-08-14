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

	topic() *topic

	setTopic(text string, author string)

	banned(c clienter) bool

	addBan(mask banMask) error

	removeBan(mask banMask) error

	names() []string

	broadcastRPL(rpl rpl, sourceID clientID, skip bool)

	broadcastCommand(cmd command, sourceID clientID, skip bool)

	modestring() string

	mode() (mode channelMode)

	addMode(mode channelMode)

	removeMode(mode channelMode)

	hasMode(mode channelMode) bool

	isInvited(c clienter) bool

	addInvite(clientID clientID)

	removeInvite(clientID clientID)

	key() string

	setKey(key string)
}

type banMask string

type channel struct {
	mu *sync.RWMutex

	n string

	t *topic

	cs      channelClientStorer
	modes   channelMode
	bans    map[banMask]bool
	invites map[clientID]bool

	o clientID

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

func (ch *channel) topic() *topic {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.t
}

func (ch *channel) names() []string {
	var names []string

	for _, c := range ch.cs.all() {
		names = append(names, c.nickname())
	}

	return names
}

func (ch *channel) broadcastRPL(rpl rpl, sourceID clientID, skip bool) {
	for _, c := range ch.cs.all() {
		if c.quitReason() != "" {
			continue
		}
		if c.id() == sourceID && skip {
			continue
		}
		c.send(rpl.rpl())
	}
}

func (ch *channel) broadcastCommand(cmd command, sourceID clientID, skip bool) {
	for _, c := range ch.cs.all() {
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

	iv, ok := ch.invites[c.id()]
	if !ok {
		return false
	}

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
