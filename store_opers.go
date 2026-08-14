package ircd

import "sync"

// OperatorStore holds operator usernames and passwords for IRC operator authentication.
type OperatorStore struct {
	mu *sync.RWMutex

	ops map[string]string
}

// NewOperatorStore returns an empty operator credential store.
func NewOperatorStore() *OperatorStore {
	return &OperatorStore{
		mu:  &sync.RWMutex{},
		ops: make(map[string]string),
	}
}

func (os *OperatorStore) add(user string, password string) {
	os.mu.Lock()
	os.ops[user] = password
	os.mu.Unlock()
}

func (os *OperatorStore) auth(user string, password string) bool {
	os.mu.RLock()
	defer os.mu.RUnlock()
	p, ok := os.ops[user]
	if !ok {
		return false
	}
	if p == password {
		return true
	}
	return false
}
