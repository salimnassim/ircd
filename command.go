package ircd

import "fmt"

type command interface {
	command() string
}

type partCommand struct {
	prefix  string
	channel string
	text    string
}

func (cmd partCommand) command() string {
	return fmt.Sprintf(
		":%s PART %s :%s",
		cmd.prefix, cmd.channel, cmd.text,
	)
}
