package ircd

import (
	"fmt"
)

type Message struct {
	Raw     string
	Tags    map[string]interface{}
	Prefix  string
	Command string
	Params  []string
}

func numericMessage(prefix string, numeric string, nickname string, line string) string {
	return fmt.Sprintf(":%s %s %s :%s\r\n", prefix, numeric, nickname, line)
}
