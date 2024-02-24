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
	if cmd.text == "" {
		return fmt.Sprintf(
			":%s PART %s :No reason given",
			cmd.prefix, cmd.channel,
		)
	}
	return fmt.Sprintf(
		":%s PART %s :%s",
		cmd.prefix, cmd.channel, cmd.text,
	)
}

type privmsgCommand struct {
	prefix string
	target string
	text   string
}

func (cmd privmsgCommand) command() string {
	return fmt.Sprintf(
		":%s PRIVMSG %s :%s",
		cmd.prefix, cmd.target, cmd.text,
	)
}

type noticeCommand struct {
	client  string
	message string
}

func (cmd noticeCommand) command() string {
	return fmt.Sprintf(
		"NOTICE %s :%s",
		cmd.client, cmd.message,
	)
}

// type errorCommand struct {
// 	text string
// }

// func (cmd errorCommand) command() string {
// 	return fmt.Sprintf(
// 		"ERROR :%s",
// 		cmd.text,
// 	)
// }

type pingCommand struct {
	text string
}

func (cmd pingCommand) command() string {
	return fmt.Sprintf(
		"PING %s",
		cmd.text,
	)
}

type modeCommand struct {
	source     string
	target     string
	modestring string
	args       string
}

func (cmd modeCommand) command() string {
	if cmd.source == "" {
		return fmt.Sprintf(
			"MODE %s %s %s",
			cmd.target, cmd.modestring, cmd.args,
		)
	}
	return fmt.Sprintf(
		":%s MODE %s %s %s",
		cmd.source, cmd.target, cmd.modestring, cmd.args,
	)
}

type joinCommand struct {
	prefix  string
	channel string
}

func (cmd joinCommand) command() string {
	return fmt.Sprintf(
		":%s JOIN %s",
		cmd.prefix, cmd.channel,
	)
}
