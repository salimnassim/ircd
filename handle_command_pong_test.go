package ircd

import "testing"

func TestCommandPong(t *testing.T) {
	got := false
	want := true

	c := newMockClient(true)
	s := NewServer(ServerConfig{
		Name: "server",
	})

	m := message{
		raw: "PING abcd12345",
	}

	handlePong(s, c, m)

	got = c.messagesPong[0]

	if got != want {
		t.Errorf("got: %t, want %t", got, want)
	}
}
