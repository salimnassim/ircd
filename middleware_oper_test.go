package ircd

import (
	"slices"
	"testing"
)

func TestMiddlewareOper(t *testing.T) {
	s := NewServer(ServerConfig{Name: "server"})
	c := newMockClient(false)
	m := message{}

	want := []string{"481 mocknick :Permission Denied - You're not an IRC operator."}
	middlewareNeedOper(s, c, m, func(s *server, c clienter, m message) {})
	if slices.Compare(c.messagesOut, want) != 0 {
		t.Errorf("got: %v, want: %v", c.messagesOut, want)
	}

	c.reset()
	c.addMode(modeClientOperator)

	want = []string{}
	middlewareNeedOper(s, c, m, func(s *server, c clienter, m message) {})
	if slices.Compare(c.messagesOut, want) != 0 {
		t.Errorf("got: %v, want: %v", c.messagesOut, want)
	}
}
