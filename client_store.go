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
}

type ClientStore struct {
	mutex   *sync.RWMutex
	clients map[*Client]bool
}

// Creates a new client store
func NewClientStore() *ClientStore {
	return &ClientStore{
		mutex:   &sync.RWMutex{},
		clients: make(map[*Client]bool),
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

	cs.clients[client] = true
}

func (cs *ClientStore) Remove(client *Client) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.clients, client)
}

func (cs *ClientStore) Get(client *Client) (*Client, bool) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	for v := range cs.clients {
		if v == client {
			return v, true
		}
	}

	return nil, false
}

func (cs *ClientStore) GetByNickname(nickname string) (*Client, bool) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	for v := range cs.clients {
		if v.GetNickname() == nickname {
			return v, true
		}
	}

	return nil, false
}
