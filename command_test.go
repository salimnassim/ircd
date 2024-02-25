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
		{
			input: partCommand{
				prefix:  "nick!user@host.fqdn",
				channel: "#testing",
				text:    "",
			},
			want: ":nick!user@host.fqdn PART #testing :No reason given",
		},
		{
			input: privmsgCommand{
				prefix: "nick!user@host.fqdn",
				target: "#testing",
				text:   "hey",
			},
			want: ":nick!user@host.fqdn PRIVMSG #testing :hey",
		},
		{
			input: noticeCommand{
				client:  "client",
				message: "hey",
			},
			want: "NOTICE client :hey",
		},
		{
			input: pingCommand{
				text: "12345",
			},
			want: "PING 12345",
		},
		{
			input: modeCommand{
				source:     "",
				target:     "client",
				modestring: "+v",
				args:       "",
			},
			want: "MODE client +v ",
		},
		{
			input: modeCommand{
				source:     "server",
				target:     "client",
				modestring: "+v",
				args:       "",
			},
			want: ":server MODE client +v ",
		},
		{
			input: joinCommand{
				prefix:  "nick!user@host.fqdn",
				channel: "#testing",
			},
			want: ":nick!user@host.fqdn JOIN #testing",
		},
	}

	for _, tc := range tcs {
		if tc.input.command() != tc.want {
			t.Errorf("got %s, want %s", tc.input.command(), tc.want)
		}
	}
}
