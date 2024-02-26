package ircd

import (
	"testing"
)

func TestMask(t *testing.T) {
	type tc struct {
		input string
		mask  string
		want  bool
	}

	tcs := []tc{
		{
			input: "nick!user@host.com",
			mask:  "nick!*@host.com",
			want:  true,
		},
		{
			input: "nick!user@host.com",
			mask:  "ni?k!user@host.com",
			want:  true,
		},
		{
			input: "nick!user@host.com",
			mask:  "nick!user@nope.com",
			want:  false,
		},
		{
			input: "nick!user@host.com",
			mask:  "nick!user@nopenopenope.com",
			want:  false,
		},
		{
			input: "nick!user@nopenopenope.com",
			mask:  "nick!user@host.com",
			want:  false,
		},
		{
			input: "nick!user@host.com",
			mask:  "nick!user@*.com",
			want:  true,
		},
		{
			input: "ni[k!user@host.com",
			mask:  "ni?k!user@*.com",
			want:  true,
		},
		{
			input: "ni[k!user@host.com",
			mask:  "ni[k!user@host.com",
			want:  true,
		},
		{
			input: "nick!user@host.com",
			mask:  "nick!asdf@*.com",
			want:  false,
		},
		{
			input: "nick!user@nopenopenope.com",
			mask:  "",
			want:  true,
		},
	}

	for _, tc := range tcs {
		parsed, err := parseMask(tc.mask)
		if err != nil {
			t.Error(err)
		}
		ok := matchMask(parsed, tc.input)
		if ok != tc.want {
			t.Errorf("got %t, want %t (mask: %s)", ok, tc.want, tc.mask)
		}
	}
}

func TestBadMaskCharacter(t *testing.T) {
	type tc struct {
		input string
		want  error
	}

	tcs := []tc{
		{
			input: string([]rune{0x00, 0x80}),
			want:  errorBadMaskCharadcter,
		},
	}

	for _, tc := range tcs {
		_, err := parseMask(tc.input)
		if err != tc.want {
			t.Errorf("got %s, want %s", err, tc.want)
		}
	}

}
