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
		{input: "PING", want: message{command: "PING"}},
		{input: "PING 12345", want: message{command: "PING", params: []string{"12345"}}},
		{input: "PING LAG206400570", want: message{command: "PING", params: []string{"LAG206400570"}}},
		{input: "version", want: message{command: "VERSION"}},
		{input: "CAP LS", want: message{command: "CAP", params: []string{"LS"}}},
		{input: "NICK salami", want: message{command: "NICK", params: []string{"salami"}}},
		{input: "USER salami salami localhost :salami", want: message{command: "USER"}},
		{input: "PONG ircd", want: message{command: "PONG", params: []string{"ircd"}}},
		{input: "JOIN #foo", want: message{command: "JOIN", params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost JOIN #foo", want: message{command: "JOIN", prefix: "salami1!salami@localhost", params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost PART #foo", want: message{command: "PART", prefix: "salami1!salami@localhost", params: []string{"#foo"}}},
		{input: ":salami1!salami@localhost PART #foo #baz", want: message{command: "PART", prefix: "salami1!salami@localhost", params: []string{"#foo", "#baz"}}},
		{input: "PRIVMSG #test :hey", want: message{command: "PRIVMSG", params: []string{"#test", "hey"}}},
		{input: "lusers", want: message{command: "LUSERS"}},
		{input: "PRIVMSG 123 :\u0001PING 1688102122 530516\u0001", want: message{command: "PRIVMSG"}},
		{input: "MODE salami +i", want: message{command: "MODE", params: []string{"salami", "+i"}}},
		{input: "MODE salami -i", want: message{command: "MODE", params: []string{"salami", "-i"}}},
		{input: "MODE #testing2", want: message{command: "MODE", params: []string{"#testing2"}}},
		{input: "WHO salami", want: message{command: "WHO", params: []string{"salami"}}},
		{input: "WHO #test", want: message{command: "WHO", params: []string{"#test"}}},
		{input: "QUIT :reason", want: message{command: "QUIT", params: []string{"reason"}}},
		{input: "QUIT :reason here", want: message{command: "QUIT", params: []string{"reason", "here"}}},
		{input: "AWAY :brb afk", want: message{command: "AWAY", params: []string{"brb", "afk"}}},
		{input: "", want: message{command: ""}},
	}

	for _, tc := range tests {
		got, err := parseMessage(tc.input)
		if err != nil {
			t.Error(err)
		}
		if got.command != tc.want.command {
			t.Errorf("got: %s, want: %s", got.command, tc.want.command)

			if len(tc.want.params) > 0 {
				for idx, val := range tc.want.params {
					if got.params[idx] != val {
						t.Errorf("got: %s, want: %s", val, tc.want.params[idx])
					}
				}
			}
		}
	}
}
