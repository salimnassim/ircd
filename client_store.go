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
	clients sync.Map
}

// Creates a new client store
func NewClientStore() *ClientStore {
	return &ClientStore{
		clients: sync.Map{},
	}
}

// Returns the size of store
func (cs *ClientStore) Size() int {
	size := 0

	cs.clients.Range(func(key, value any) bool {
		size++
		return true
	})

	return size
}

// Adds client to store
func (cs *ClientStore) Add(client *Client) {
	cs.clients.Store(client.id, client)
}

// Removes client from store
func (cs *ClientStore) Remove(client *Client) {
	cs.clients.Delete(client.id)
}

func (cs *ClientStore) GetByNickname(nickname string) (*Client, bool) {
	var client *Client

	cs.clients.Range(func(key, value any) bool {
		if value.(*Client).Nickname() == nickname {
			client = value.(*Client)
			return false
		}
		return true
	})

	if client == nil {
		return nil, false
	}

	return client, true
}

func (cs *ClientStore) Whois(nickname string) (clientWhois, bool) {
	var who *Client

	who, _ = cs.GetByNickname(nickname)

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
