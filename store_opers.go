package ircd

import "sync"

// OperatorStore holds operator usernames and passwords for IRC operator authentication.
type OperatorStore struct {
	mu sync.RWMutex

	ops map[string]string
}

// NewOperatorStore returns an empty operator credential store.
func NewOperatorStore() *OperatorStore {
	return &OperatorStore{
		ops: make(map[string]string),
	}
}

func (o *OperatorStore) add(user string, password string) {
	o.mu.Lock()
	o.ops[user] = password
	o.mu.Unlock()
}

func (o *OperatorStore) auth(user string, password string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	p, ok := o.ops[user]
	if !ok {
		return false
	}
	if p == password {
		return true
	}
	return false
}
