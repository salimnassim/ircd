package ircd

import (
	"testing"
)

func TestCommandAway(t *testing.T) {
	got := ""
	s := NewServer(ServerConfig{
		Name: "server",
	})
	c := newMockClient(false)

	t.Run("set away", func(t *testing.T) {
		m := message{
			command: "AWAY",
			params:  []string{"reason"},
		}
		want := "reason"
		handleAway(s, c, m)

		if c.away() != want {
			t.Errorf("got: %s, want: %s", got, want)
		}
	})

	t.Run("set unaway", func(t *testing.T) {
		m2 := message{
			command: "AWAY",
		}
		want := ""
		handleAway(s, c, m2)
		if c.away() != "" {
			t.Errorf("got: %s, want: %s", got, want)
		}
	})

}
