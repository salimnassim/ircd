package ircd

import (
	"errors"
	"sync"
)

var errNickInUse = errors.New("nickname is already in use")

type clientID string

type clientDirectory struct {
	mu sync.Mutex

	byID        map[clientID]*sessionHandle
	byNick      map[string]clientID
	reverseNick map[clientID]string
}

func newClientDirectory() *clientDirectory {
	return &clientDirectory{
		byID:        make(map[clientID]*sessionHandle),
		byNick:      make(map[string]clientID),
		reverseNick: make(map[clientID]string),
	}
}

func (d *clientDirectory) register(h *sessionHandle) {
	d.mu.Lock()
	d.byID[h.id] = h
	d.mu.Unlock()
}

func (d *clientDirectory) claimNick(id clientID, nick string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if owner, ok := d.byNick[nick]; ok && owner != id {
		return errNickInUse
	}

	if old, ok := d.reverseNick[id]; ok {
		delete(d.byNick, old)
	}

	d.byNick[nick] = id
	d.reverseNick[id] = nick
	return nil
}

func (d *clientDirectory) unregister(id clientID) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.byID, id)
	if nick, ok := d.reverseNick[id]; ok {
		delete(d.byNick, nick)
		delete(d.reverseNick, id)
	}
}

func (d *clientDirectory) lookupNick(nick string) (*sessionHandle, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	id, ok := d.byNick[nick]
	if !ok {
		return nil, false
	}
	h, ok := d.byID[id]
	return h, ok
}

func (d *clientDirectory) lookupID(id clientID) (*sessionHandle, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	h, ok := d.byID[id]
	return h, ok
}

func (d *clientDirectory) all() []*sessionHandle {
	d.mu.Lock()
	defer d.mu.Unlock()

	handles := make([]*sessionHandle, 0, len(d.byID))
	for _, h := range d.byID {
		handles = append(handles, h)
	}
	return handles
}

func (d *clientDirectory) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.byID)
}
