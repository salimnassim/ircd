package ircd

func handlePong(s *server, c *client, m message) {
	c.gotPong <- true
}
