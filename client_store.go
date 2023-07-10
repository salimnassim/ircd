package ircd

import (
	"sync"
)

type ClientStoreable interface {
	Size() int
	Add(*Client)
	Remove(*Client)
	GetByNickname(string) (*Client, bool)
	Whois(string) (clientWhois, bool)
}

type clientWhois struct {
	nickname string
	username string
	realname string
	hostname string
	channels []string
}

type ClientStore struct {
	mutex   *sync.RWMutex
	clients map[string]*Client
}

// Creates a new client store
func NewClientStore() *ClientStore {
	return &ClientStore{
		mutex:   &sync.RWMutex{},
		clients: map[string]*Client{},
	}
}

// Returns the size of store
func (cs *ClientStore) Size() int {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	size := len(cs.clients)
	return size
}

// Adds client to store
func (cs *ClientStore) Add(client *Client) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.clients[client.id] = client
}

// Removes client from store
func (cs *ClientStore) Remove(client *Client) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.clients, client.id)
}

func (cs *ClientStore) GetByNickname(nickname string) (*Client, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	for _, c := range cs.clients {
		c.mu.RLock()
		if c.nickname == nickname {
			c.mu.RUnlock()
			return c, true
		}
		c.mu.RUnlock()
	}

	return nil, false
}

func (cs *ClientStore) Whois(nickname string) (clientWhois, bool) {
	var who *Client

	cs.mutex.RLock()
	for _, c := range cs.clients {
		if c.nickname == nickname {
			who = c
		}
	}
	cs.mutex.RUnlock()

	if who == nil {
		return clientWhois{}, false
	}

	who.mu.Lock()
	whois := clientWhois{
		nickname: who.nickname,
		username: who.username,
		realname: who.realname,
		hostname: who.hostname,
		channels: []string{},
	}
	who.mu.Unlock()

	return whois, true
}
