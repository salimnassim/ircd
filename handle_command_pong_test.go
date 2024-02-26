package ircd

import "testing"

func TestCommandPong(t *testing.T) {
	c := newMockClient(true)
	s := NewServer(ServerConfig{
		Name: "server",
	})
	m := message{
		raw: "PING abcd12345",
	}

	got := false
	want := true
	handlePong(s, c, m)

	got = c.messagesPong[0]

	if got != want {
		t.Errorf("got: %t, want %t", got, want)
	}
}
