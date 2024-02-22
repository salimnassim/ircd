package ircd

import (
	"testing"
)

func TestParseModestring(t *testing.T) {

	type tcg struct {
		add []clientMode
		del []clientMode
	}

	type tc struct {
		input string
		want  tcg
	}

	tcs := []tc{
		{
			input: "+ir-w",
			want: tcg{
				add: []clientMode{
					modeClientInvisible,
					modeClientRegistered,
				},
				del: []clientMode{
					modeClientWallops,
				},
			},
		},
	}

	for _, tc := range tcs {
		a, d := parseModestring[clientMode](tc.input, clientModeMap)

		for i, v := range a {
			if v != tc.want.add[i] {
				t.Errorf("got: %d, want: %d", v, tc.want.add[i])
			}
		}

		for i, v := range d {
			if v != tc.want.del[i] {
				t.Errorf("got: %d, want: %d", v, tc.want.add[i])
			}
		}
	}
}
