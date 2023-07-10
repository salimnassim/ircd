package ircd

import (
	"sync"
)

type ClientStoreable interface {
	Size() int
	Add(*Client)
	Remove(*Client)
	Get(*Client) (*Client, bool)
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

func (cs *ClientStore) Size() int {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	len := len(cs.clients)
	return len
}

func (cs *ClientStore) Add(client *Client) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.clients[client.id] = client
}

func (cs *ClientStore) Remove(client *Client) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.clients, client.id)
}

func (cs *ClientStore) Get(client *Client) (*Client, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	for id, c := range cs.clients {
		if id == client.id {
			return c, true
		}
	}

	return nil, false
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
	var ptr *Client

	cs.mutex.RLock()
	for _, c := range cs.clients {
		if c.nickname == nickname {
			ptr = c
		}
	}
	cs.mutex.RUnlock()

	if ptr == nil {
		return clientWhois{}, false
	}

	ptr.mu.Lock()
	whois := clientWhois{
		nickname: ptr.nickname,
		username: ptr.username,
		realname: ptr.realname,
		hostname: ptr.hostname,
		channels: []string{},
	}
	ptr.mu.Unlock()

	return whois, true
}
