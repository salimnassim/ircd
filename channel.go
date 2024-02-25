package ircd

import (
	"fmt"
	"sync"
	"time"
)

type channel struct {
	mu       *sync.RWMutex
	name     string
	t        *topic
	clients  ChannelClientStorer
	modes    channelMode
	owner    clientID
	password string
	secret   bool
}

type topic struct {
	text      string
	timestamp int
	author    string
}

func newChannel(channelName string, owner clientID) *channel {
	channel := &channel{
		mu:   &sync.RWMutex{},
		name: channelName,
		t: &topic{
			text:      "",
			timestamp: 0,
			author:    "",
		},
		clients:  newChannelClientStore(),
		modes:    0,
		owner:    owner,
		password: "",
		secret:   false,
	}

	return channel
}

// Sets channel topic.
func (ch *channel) setTopic(text string, author string) {
	ch.mu.Lock()
	ch.t.text = text
	ch.t.timestamp = int(time.Now().Unix())
	ch.t.author = author
	ch.mu.Unlock()
}

// Returns current topic.
func (ch *channel) topic() *topic {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.t
}

// Adds client to channel. If password does not match, an error is returned.
func (ch *channel) addClient(c clienter, password string) error {
	if password != "" && ch.password != password {
		return errorBadChannelKey
	}

	ch.clients.add(c)

	return nil
}

// Remove client from channel.
func (ch *channel) removeClient(c clienter) {
	ch.clients.delete(c.id())
}

// Returns channel users delimited by a space for RPL_NAMREPLY.
func (ch *channel) names() []string {
	var names []string

	for _, c := range ch.clients.all() {
		if ch.owner == c.id() {
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
	for _, c := range ch.clients.all() {
		if c.id() == sourceID && skip {
			continue
		}
		c.send(rpl.format())
	}
}

// Send command to all clients on the channel.
// If skip is true, the client in source will not receive the message.
func (ch *channel) broadcastCommand(cmd command, sourceID clientID, skip bool) {
	for _, c := range ch.clients.all() {
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
