package ircd

import "strings"

type message struct {
	raw     string
	tags    map[string]any
	prefix  string
	command string
	params  []string
}

func (m message) isTargetChannel() bool {
	target := ""
	if len(m.params) >= 1 {
		target = m.params[0]
	} else {
		return false
	}
	if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
		return true
	}
	return false
}
