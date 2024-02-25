package ircd

// Require client to be an operator for commands using this middleware.
func middlewareNeedOper(s *server, c clienter, m message, next handlerFunc) handlerFunc {
	if !c.hasMode(modeClientOperator) {
		c.sendRPL(s.name, errNoPrivileges{
			client: c.nickname(),
		})
		return nil
	}
	return next
}
