package ircd

func middlewareNeedParams(count int) middlewareFunc {
	return func(s *server, c clienter, m message, next handlerFunc) handlerFunc {
		if len(m.params) < count {
			c.sendRPL(s.name, errNeedMoreParams{
				client:  c.nickname(),
				command: m.command,
			})
			return nil
		}
		return next
	}
}
