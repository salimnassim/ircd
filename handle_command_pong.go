package ircd

func handlePong(s *server, c clienter, m message) {
	c.pong(true)
}
