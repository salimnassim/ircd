package ircd

import (
	"slices"
	"testing"
)

func TestMiddlewareHandshake(t *testing.T) {
	s := NewServer(ServerConfig{Name: "server"})
	c := newMockClient(false)

	m := message{}

	want := []string{"451 mocknick :You have not registered."}
	middlewareNeedHandshake(s, c, m, func(s *server, c clienter, m message) {})
	if slices.Compare(c.messagesOut, want) != 0 {
		t.Errorf("got: %v, want: %v", c.messagesOut, want)
	}

	c.reset()
	c.setHandshake(true)

	want = []string{}
	middlewareNeedHandshake(s, c, m, func(s *server, c clienter, m message) {})
	if slices.Compare(c.messagesOut, want) != 0 {
		t.Errorf("got: %v, want: %v", c.messagesOut, want)
	}
}
