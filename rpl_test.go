package ircd

import "testing"

func TestRPLNamreply(t *testing.T) {
	type tc struct {
		want  string
		input rpl
	}

	tcs := []tc{
		{
			want: "251 client :There are 1 users and 2 invisible on 3 servers",
			input: rplLuserClient{
				client:    "client",
				users:     1,
				invisible: 2,
				servers:   3,
			},
		},
		{
			want: "252 client 4 :operator(s) online",
			input: rplLuserOp{
				client: "client",
				ops:    4,
			},
		},
		{
			want: "254 client 5 :channels formed",
			input: rplLuserChannels{
				client:   "client",
				channels: 5,
			},
		},
		{
			want: "353 client = testing :~foo @baz qax",
			input: rplNamReply{
				client:  "client",
				symbol:  "=",
				channel: "testing",
				nicks: []string{
					"~foo", "@baz", "qax",
				},
			},
		},
		{
			want: "432 client nickname :Erroneus nickname.",
			input: rplErroneusNickname{
				client: "client",
				nick:   "nickname",
			},
		},
		{
			want: "433 client nickname :Nickname is already in use.",
			input: rplNicknameInUse{
				client: "client",
				nick:   "nickname",
			},
		},
		{
			want: "461 client WHO :Not enough parameters.",
			input: rplNeedMoreParams{
				client:  "client",
				command: "WHO",
			},
		},
	}

	for _, tc := range tcs {
		m := tc.input.format()

		if m != tc.want {
			t.Errorf("got %s, want %s", m, tc.want)
		}
	}

}
