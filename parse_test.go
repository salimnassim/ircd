package ircd

import (
	"testing"
)

type test struct {
	input string
	want  message
}

func TestParse(t *testing.T) {
	tests := []test{
		{input: "PING", want: message{Command: "PING"}},
		{input: "PING 12345", want: message{Command: "PING", Params: []string{"12345"}}},
		{input: "PING LAG206400570", want: message{Command: "PING", Params: []string{"LAG206400570"}}},
		{input: "version", want: message{Command: "VERSION"}},
		{input: "CAP LS", want: message{Command: "CAP", Params: []string{"LS"}}},
		{input: "NICK salami", want: message{Command: "NICK", Params: []string{"salami"}}},
		{input: "USER salami salami localhost :salami", want: message{Command: "USER"}},
		{input: "PONG ircd", want: message{Command: "PONG", Params: []string{"ircd"}}},
		{input: "JOIN #foo", want: message{Command: "JOIN", Params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost JOIN #foo", want: message{Command: "JOIN", Prefix: "salami1!salami@localhost", Params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost PART #foo", want: message{Command: "PART", Prefix: "salami1!salami@localhost", Params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost PART #foo #baz", want: message{Command: "PART", Prefix: "salami1!salami@localhost", Params: []string{"#foo", "#baz"}}},
		{input: "PRIVMSG #test :hey", want: message{Command: "PRIVMSG", Params: []string{"#test", "hey"}}},
		{input: "lusers", want: message{Command: "LUSERS"}},
		{input: "PRIVMSG 123 :\u0001PING 1688102122 530516\u0001", want: message{Command: "PRIVMSG"}},
		{input: "MODE salami +i", want: message{Command: "MODE", Params: []string{"salami", "+i"}}},
		{input: "MODE salami -i", want: message{Command: "MODE", Params: []string{"salami", "-i"}}},
		{input: "WHO salami", want: message{Command: "WHO", Params: []string{"salami"}}},
		{input: "WHO #test", want: message{Command: "WHO", Params: []string{"#test"}}},
		{input: "", want: message{Command: ""}},
	}

	for _, tc := range tests {
		got, err := parseMessage(tc.input)
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
