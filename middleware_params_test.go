package ircd

import (
	"slices"
	"testing"
)

func TestMiddlewareParams(t *testing.T) {
	s := NewServer(ServerConfig{Name: "server"})
	c := newMockClient(false)
	m := message{command: "TEST", params: []string{"one"}}

	r := NewCommandRouter(s)
	r.registerHandler("TEST", func(s *server, c clienter, m message) {}, middlewareNeedParams(2))
	r.handle(s, c, m)

	t.Run("not enough params", func(t *testing.T) {
		want := []string{"461 mocknick TEST :Not enough parameters."}
		if slices.Compare(c.messagesOut, want) != 0 {
			t.Errorf("got %v, want %v", c.messagesOut, want)
		}
	})

	c.reset()

	r.registerHandler("TEST", func(s *server, c clienter, m message) {}, middlewareNeedParams(1))
	r.handle(s, c, m)

	t.Run("enough params", func(t *testing.T) {
		want := []string{}
		if slices.Compare(c.messagesOut, want) != 0 {
			t.Errorf("got %v, want %v", c.messagesOut, want)
		}
	})
}
