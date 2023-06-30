package ircd_test

import (
	"fmt"
	"testing"

	"github.com/salimnassim/ircd"
)

type test struct {
	input string
	want  ircd.Message
}

func TestMessageParse(t *testing.T) {

	tests := []test{
		{input: "PING", want: ircd.Message{Command: "PING"}},
		{input: "PING 12345", want: ircd.Message{Command: "PING", Params: []string{"12345"}}},
		{input: "version", want: ircd.Message{Command: "VERSION"}},
		{input: "CAP LS", want: ircd.Message{Command: "CAP"}},
		{input: "NICK salami", want: ircd.Message{Command: "NICK"}},
		{input: "USER salami salami localhost :salami", want: ircd.Message{Command: "USER"}},
		{input: "PONG ircd", want: ircd.Message{Command: "PONG"}},
		{input: "JOIN #foo", want: ircd.Message{Command: "JOIN"}},
		{input: ":salami1!salami@localhost JOIN #foo", want: ircd.Message{Command: "JOIN"}},
		{input: ":salami1!salami@localhost PART #foo", want: ircd.Message{Command: "PART"}},
		{input: "PRIVMSG #test :hey", want: ircd.Message{Command: "PRIVMSG"}},
		{input: "lusers", want: ircd.Message{Command: "LUSERS"}},
		{input: "PRIVMSG 123 :\u0001PING 1688102122 530516\u0001", want: ircd.Message{Command: "PRIVMSG"}},
	}

	for _, tc := range tests {
		got, err := ircd.Parse(tc.input)
		if err != nil {
			t.Error(err)
		}
		if got.Command != tc.want.Command {
			t.Errorf("got: %s, want: %s", got.Command, tc.want.Command)

			if len(tc.want.Params) > 0 {
				for idx, val := range tc.want.Params {
					if got.Params[idx] != val {
						fmt.Printf("got: %s, want: %s", val, tc.want.Params[idx])
						t.Errorf("got: %s, want: %s", val, tc.want.Params[idx])
					}
				}
			}
		}
	}
}
