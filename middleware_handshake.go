package ircd

// Require client handhake for commands using this middleware.
func middlewareNeedHandshake(s *server, c *client, m message, next handlerFunc) handlerFunc {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return nil
	}
	return next
}
