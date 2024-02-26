package ircd

import (
	"slices"
	"testing"
)

func TestCommandOper(t *testing.T) {
	var want []string

	c := newMockClient(true)
	s := NewServer(ServerConfig{
		Name: "server",
	})
	s.Operators.add("test", "test")

	m := message{
		command: "OPER",
		params:  []string{"wrong", "wrong"},
	}

	m2 := message{
		command: "OPER",
		params:  []string{"test", "test"},
	}

	want = []string{"464 mocknick :Password incorrect."}
	handleOper(s, c, m)
	if slices.Compare(c.messagesOut, want) != 0 {
		t.Errorf("got %v, want: %v", c.messagesOut, want)
	}

	c.reset()

	want = []string{"MODE mocknick +o", "381 mocknick :You are now an IRC operator."}
	handleOper(s, c, m2)
	if slices.Compare(c.messagesOut, want) != 0 {
		t.Errorf("got %v, want: %v", c.messagesOut, want)
	}
}
