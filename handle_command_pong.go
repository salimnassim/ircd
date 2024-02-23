package ircd

func handlePong(s *server, c *client, m message) {
	c.pong <- true
}
