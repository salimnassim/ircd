package ircd

import (
	"testing"
)

func TestCommandAway(t *testing.T) {
	got := ""

	s := NewServer(ServerConfig{
		Name: "server",
	})

	c := &clientMock{}

	c.setNickname("client")

	want := "reason"
	m := message{
		command: "AWAY",
		params:  []string{"reason"},
	}

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
