package ircd

import "testing"

func TestRPL(t *testing.T) {
	type tc struct {
		want  string
		input rpl
	}

	tcs := []tc{
		{
			want: "001 client :Welcome to the testing Network, client@hostname",
			input: rplWelcome{
				client:   "client",
				network:  "testing",
				hostname: "client@hostname",
			},
		},
		{
			want: "002 client :Your host is name, running version 1",
			input: rplYourHost{
				client:     "client",
				serverName: "name",
				version:    "1",
			},
		},
		{
			want: "221 client +viz",
			input: rplUModeIs{
				client:     "client",
				modestring: "+viz",
			},
		},
		{
			want: "251 client :There are 1 users (2 invisible) on 3 servers",
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
			want: "254 client 5 :channels formed.",
			input: rplLuserChannels{
				client:   "client",
				channels: 5,
			},
		},
		{
			want: "301 client nick :text",
			input: rplAway{
				client:  "client",
				nick:    "nick",
				message: "text",
			},
		},
		{
			want: "305 client :You are no longer marked as being away.",
			input: rplUnAway{
				client: "client",
			},
		},
		{
			want: "306 client :You have been marked as being away.",
			input: rplNowAway{
				client: "client",
			},
		},
		{
			want: "311 client nick username host * :realname",
			input: rplWhoisUser{
				client:   "client",
				nick:     "nick",
				username: "username",
				host:     "host",
				realname: "realname",
			},
		},
		{
			want: "315 client mask :End of WHO list.",
			input: rplEndOfWho{
				client: "client",
				mask:   "mask",
			},
		},
		{
			want: "319 client nick :#foo #baz",
			input: rplWhoisChannels{
				client:   "client",
				nick:     "nick",
				channels: []string{"#foo", "#baz"},
			},
		},
		{
			want: "320 client nick :text",
			input: rplWhoisSpecial{
				client: "client",
				nick:   "nick",
				text:   "text",
			},
		},
		{
			want: "324 client #channel +k password",
			input: rplChannelModeIs{
				client:     "client",
				channel:    "#channel",
				modestring: "+k",
				modeargs:   "password",
			},
		},
		{
			want: "331 client #channel :No topic is set.",
			input: rplNoTopic{
				client:  "client",
				channel: "#channel",
			},
		},
		{
			want: "332 client #channel :foo bar baz",
			input: rplTopic{
				client:  "client",
				channel: "#channel",
				topic:   "foo bar baz",
			},
		},
		{
			want: "333 client #channel bob 12345",
			input: rplTopicWhoTime{
				client:  "client",
				channel: "#channel",
				nick:    "bob",
				setat:   12345,
			},
		},
		{
			want: "352 client #channel username host server nick H :0 realname",
			input: rplWhoReply{
				client:   "client",
				channel:  "#channel",
				username: "username",
				host:     "host",
				server:   "server",
				nick:     "nick",
				flags:    "H",
				hopcount: 0,
				realname: "realname",
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
			want: "366 client #channel :End of /NAMES list.",
			input: rplEndOfNames{
				client:  "client",
				channel: "#channel",
			},
		},
		{
			want: "372 client :text",
			input: rplMotd{
				client: "client",
				text:   "text",
			},
		},
		{
			want: "375 client :- server text",
			input: rplMotdStart{
				client: "client",
				server: "server",
				text:   "text",
			},
		},
		{
			want: "376 client :End of /MOTD command.",
			input: rplEndOfMotd{
				client: "client",
			},
		},
		{
			want: "401 client nick :No such nickname.",
			input: errNoSuchNick{
				client: "client",
				nick:   "nick",
			},
		},
		{
			want: "403 client #channel :No such channel.",
			input: errNoSuchChannel{
				client:  "client",
				channel: "#channel",
			},
		},
		{
			want: "431 client :No nickname given.",
			input: errNoNicknameGiven{
				client: "client",
			},
		},
		{
			want: "432 client nickname :Erroneus nickname.",
			input: errErroneusNickname{
				client: "client",
				nick:   "nickname",
			},
		},
		{
			want: "433 client nickname :Nickname is already in use.",
			input: errNicknameInUse{
				client: "client",
				nick:   "nickname",
			},
		},
		{
			want: "442 client #channel :You are not on that channel.",
			input: errNotOnChannel{
				client:  "client",
				channel: "#channel",
			},
		},
		{
			want: "451 client :You have not registered.",
			input: errNotRegistered{
				client: "client",
			},
		},
		{
			want: "461 client WHO :Not enough parameters.",
			input: errNeedMoreParams{
				client:  "client",
				command: "WHO",
			},
		},
		{
			want: "462 client :You may not reregister.",
			input: errAlreadyRegistered{
				client: "client",
			},
		},
		{
			want: "474 client #channel :Cannot join channel (+z)",
			input: errBannedFromChan{
				client:  "client",
				channel: "#channel",
			},
		},
		{
			want: "475 client #channel :Bad channel key (+k).",
			input: errBadChannelKey{
				client:  "client",
				channel: "#channel",
			},
		},
		{
			want: "502 client :Can't change mode for other users.",
			input: errUsersDontMatch{
				client: "client",
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
