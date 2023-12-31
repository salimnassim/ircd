package ircd_test

import (
	"testing"

	"github.com/salimnassim/ircd"
)

type test struct {
	input string
	want  ircd.Message
}

func TestParse(t *testing.T) {

	tests := []test{
		{input: "PING", want: ircd.Message{Command: "PING"}},
		{input: "PING 12345", want: ircd.Message{Command: "PING", Params: []string{"12345"}}},
		{input: "PING LAG206400570", want: ircd.Message{Command: "PING", Params: []string{"LAG206400570"}}},
		{input: "version", want: ircd.Message{Command: "VERSION"}},
		{input: "CAP LS", want: ircd.Message{Command: "CAP", Params: []string{"LS"}}},
		{input: "NICK salami", want: ircd.Message{Command: "NICK", Params: []string{"salami"}}},
		{input: "USER salami salami localhost :salami", want: ircd.Message{Command: "USER"}},
		{input: "PONG ircd", want: ircd.Message{Command: "PONG", Params: []string{"ircd"}}},
		{input: "JOIN #foo", want: ircd.Message{Command: "JOIN", Params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost JOIN #foo", want: ircd.Message{Command: "JOIN", Prefix: "salami1!salami@localhost", Params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost PART #foo", want: ircd.Message{Command: "PART", Prefix: "salami1!salami@localhost", Params: []string{"#foo"}}},
		{input: "PRIVMSG #test :hey", want: ircd.Message{Command: "PRIVMSG", Params: []string{"#test", "hey"}}},
		{input: "lusers", want: ircd.Message{Command: "LUSERS"}},
		{input: "PRIVMSG 123 :\u0001PING 1688102122 530516\u0001", want: ircd.Message{Command: "PRIVMSG"}},
		{input: "MODE salami +i", want: ircd.Message{Command: "MODE", Params: []string{"salami", "+i"}}},
		{input: "MODE salami -i", want: ircd.Message{Command: "MODE", Params: []string{"salami", "-i"}}},
		{input: "WHO salami", want: ircd.Message{Command: "WHO", Params: []string{"salami"}}},
		{input: "WHO #test", want: ircd.Message{Command: "WHO", Params: []string{"#test"}}},
		{input: "", want: ircd.Message{Command: ""}},
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
						t.Errorf("got: %s, want: %s", val, tc.want.Params[idx])
					}
				}
			}
		}
	}
}
