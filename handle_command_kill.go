package ircd

import (
	"fmt"
	"strings"
)

type killError struct {
	byNick string
	reason string
}

func (e *killError) Error() string {
	return fmt.Sprintf("Killed by %s (%s)", e.byNick, e.reason)
}

func (e *killError) Is(target error) bool {
	return target == errKilled
}

func (s *session) cmdKill(m message) {
	if s.modes&modeClientOperator == 0 {
		s.handle.deliver(s.formatRPL(errNoPrivileges{client: s.nick}))
		return
	}

	targetNick := m.params[0]
	target, ok := s.deps.clients.lookupNick(targetNick)
	if !ok {
		s.handle.deliver(s.formatRPL(errNoSuchNick{client: s.nick, nick: targetNick}))
		return
	}

	reason := strings.Join(m.params[1:], " ")

	target.terminate(&killError{byNick: s.nick, reason: reason})
}
