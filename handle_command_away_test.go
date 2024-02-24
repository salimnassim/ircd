package ircd

import (
	"testing"
)

func TestCommandAway(t *testing.T) {
	got := ""

	s := NewServer(ServerConfig{
		Name: "server",
	})

	c, err := newClient(&connMock{}, "test")
	if err != nil {
		t.Error(err)
	}

	c.setNickname("client")

	go func() {
		for {
			select {
			case <-c.pong:
				continue
			case <-c.recv:
				continue
			case <-c.send:
				continue
			case <-c.stop:
				return
			}
		}
	}()

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
		params:  []string{""},
	}

	handleAway(s, c, m2)

	if c.away() != "" {
		t.Errorf("got: %s, want: %s", got, want)
	}

	c.stop <- "test"
}
