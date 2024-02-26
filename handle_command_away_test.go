package ircd

import (
	"testing"
)

func TestCommandAway(t *testing.T) {
	s := NewServer(ServerConfig{
		Name: "server",
	})
	c := newMockClient(false)

	m := message{
		command: "AWAY",
		params:  []string{"reason"},
	}

	got := ""
	want := "reason"
	handleAway(s, c, m)

	if c.away() != want {
		t.Errorf("got: %s, want: %s", got, want)
	}

	want = ""
	m2 := message{
		command: "AWAY",
	}

	handleAway(s, c, m2)

	if c.away() != "" {
		t.Errorf("got: %s, want: %s", got, want)
	}
}
