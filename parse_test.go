package ircd_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/salimnassim/ircd"
)

type test struct {
	input string
	want  []string
}

func Test(t *testing.T) {

	tests := []test{
		{input: "version", want: []string{"a/b/c"}},
		{input: "CAP LS", want: []string{"a", "b", "c"}},
		{input: "NICK salami", want: []string{"a/b/c"}},
		{input: "USER salami salami localhost :salami", want: []string{"abc"}},
		{input: "PONG ircd", want: []string{"a/b/c"}},
		{input: "JOIN #foo", want: []string{"a/b/c"}},
		{input: ":salami1!salami@localhost JOIN #foo", want: []string{"a/b/c"}},
		{input: ":salami1!salami@localhost PART #foo", want: []string{"a/b/c"}},
		{input: "PRIVMSG #test :hey", want: []string{"a/b/c"}},
		{input: "lusers", want: []string{"a/b/c"}},
	}

	for _, tc := range tests {
		got, err := ircd.Parse(tc.input)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("got: %s, %s, %s\n\n", got.Command, got.Prefix, strings.Join(got.Params, ";"))
	}
}
