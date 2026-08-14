package ircd

func middlewareNeedOper(s *server, c clienter, _ message, next handlerFunc) handlerFunc {
	if !c.hasMode(modeClientOperator) {
		c.sendRPL(s.name, errNoPrivileges{
			client: c.nickname(),
		})
		return nil
	}
	return next
}
