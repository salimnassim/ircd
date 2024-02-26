package ircd

import "testing"

func TestCommandPing(t *testing.T) {
	s := NewServer(ServerConfig{
		Name: "server",
	})
	c := &clientMock{}
	m := message{
		raw: "PING abcd12345",
	}

	got := ""
	want := "PONG abcd12345"
	handlePing(s, c, m)

	got = c.messagesOut[0]

	if got != want {
		t.Errorf("got: %s, want %s", got, want)
	}
}
