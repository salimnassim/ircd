package ircd

func handleLusers(s *server, c clienter, m message) {
	visible, invisible, channels := s.Stats()

	c.sendRPL(s.name, rplLuserClient{
		client:    c.nickname(),
		users:     (visible + invisible),
		invisible: invisible,
		servers:   1,
	})
	c.sendRPL(s.name, rplLuserOp{
		client: c.nickname(),
		ops:    0,
	})
	c.sendRPL(s.name, rplLuserChannels{
		client:   c.nickname(),
		channels: channels,
	})
}
