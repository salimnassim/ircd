package ircd

import (
	"context"
	"errors"
	"sync"
)

var errChannelNotFound = errors.New("channel not found")

type channelDirectory struct {
	mu sync.Mutex

	byName  map[string]*channelActor
	pending map[string]int
}

func newChannelDirectory() *channelDirectory {
	return &channelDirectory{
		byName:  make(map[string]*channelActor),
		pending: make(map[string]int),
	}
}

type channelActorDeps struct {
	serverName string
	clients    *clientDirectory
}

func (d *channelDirectory) dispatch(ctx context.Context, deps channelActorDeps, name string, ev channelEvent, createIfMissing bool, owner clientID) error {
	d.mu.Lock()

	ch, exists := d.byName[name]
	if !exists {
		if !createIfMissing {
			d.mu.Unlock()
			return errChannelNotFound
		}
		ch = newChannelActor(name, owner, d, deps)
		d.byName[name] = ch
		go ch.run(ctx)
	}

	d.pending[name]++
	d.mu.Unlock()

	ch.inbox <- ev
	return nil
}

func (d *channelDirectory) doneProcessing(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.pending[name] <= 1 {
		delete(d.pending, name)
		return
	}
	d.pending[name]--
}

func (d *channelDirectory) removeIfIdle(name string, ch *channelActor) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.byName[name] != ch {
		return false
	}
	if d.pending[name] > 0 {
		return false
	}
	delete(d.byName, name)
	delete(d.pending, name)
	return true
}

func (d *channelDirectory) get(name string) (*channelActor, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	ch, ok := d.byName[name]
	return ch, ok
}

func (d *channelDirectory) all() []*channelActor {
	d.mu.Lock()
	defer d.mu.Unlock()

	channels := make([]*channelActor, 0, len(d.byName))
	for _, ch := range d.byName {
		channels = append(channels, ch)
	}
	return channels
}

func (d *channelDirectory) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.byName)
}

func (d *channelDirectory) snapshot(name string) (channelSnapshot, bool) {
	ch, ok := d.get(name)
	if !ok {
		return channelSnapshot{}, false
	}
	snap := ch.snapshot.Load()
	if snap == nil {
		return channelSnapshot{}, false
	}
	return *snap, true
}
