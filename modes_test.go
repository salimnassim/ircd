package ircd

import (
	"slices"
	"testing"
)

func TestParseClientModestring(t *testing.T) {
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

func TestParseChannelModestring(t *testing.T) {
	type tcg struct {
		add []channelMode
		del []channelMode
	}

	type tc struct {
		input string
		want  tcg
	}

	tcs := []tc{
		{
			input: "+z",
			want: tcg{
				add: []channelMode{
					modeChannelTLSOnly,
				},
				del: []channelMode{},
			},
		},
	}

	for _, tc := range tcs {
		a, d := parseModestring[channelMode](tc.input, channelModeMap)

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

func TestDiffModes(t *testing.T) {

	type tcg struct {
		old clientMode
		new clientMode
	}

	type tcd struct {
		add []clientMode
		del []clientMode
	}

	type tc struct {
		input tcg
		want  tcd
	}

	tcs := []tc{
		{
			input: tcg{
				old: modeClientInvisible | modeClientVhost,
				new: modeClientVhost,
			},
			want: tcd{
				add: []clientMode{},
				del: []clientMode{modeClientInvisible},
			},
		},
		{
			input: tcg{
				old: modeClientInvisible,
				new: modeClientInvisible | modeClientOperator | modeClientRegistered,
			},
			want: tcd{
				add: []clientMode{modeClientOperator, modeClientRegistered},
				del: []clientMode{},
			},
		},
	}

	for _, tc := range tcs {
		a, d := diffModes[clientMode](tc.input.old, tc.input.new, clientModeMap)
		if slices.Compare(a, tc.want.add) != 0 {
			t.Errorf("add slices do not match")
		}
		if slices.Compare(d, tc.want.del) != 0 {
			t.Errorf("del slices do not match")
		}
	}

}
