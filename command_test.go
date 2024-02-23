package ircd

import "testing"

func TestCommands(t *testing.T) {
	type tc struct {
		input command
		want  string
	}

	tcs := []tc{
		{
			input: partCommand{
				prefix:  "nick!user@host.fqdn",
				channel: "#testing",
				text:    "i am parting",
			},
			want: ":nick!user@host.fqdn PART #testing :i am parting",
		},
	}

	for _, tc := range tcs {
		if tc.input.command() != tc.want {
			t.Errorf("got %s, want %s", tc.input.command(), tc.want)
		}
	}
}
