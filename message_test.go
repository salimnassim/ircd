package ircd

import "testing"

func TestIsTargetChannel(t *testing.T) {
	type tc struct {
		input message
		want  bool
	}

	tcs := []tc{
		{
			input: message{params: []string{"#testing"}},
			want:  true,
		},
		{
			input: message{params: []string{"&foo"}},
			want:  true,
		},
		{
			input: message{params: []string{"foo"}},
			want:  false,
		},
		{
			input: message{},
			want:  false,
		},
	}

	for _, tc := range tcs {
		if tc.input.isTargetChannel() != tc.want {
			t.Errorf("got %t, want: %t", tc.input.isTargetChannel(), tc.want)
		}
	}
}
