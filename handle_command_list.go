package ircd

func handleList(s *server, c clienter, m message) {
	// if no params, list all channels
	if len(m.params) == 0 {
		c.sendRPL(s.name, rplListStart{
			client: c.nickname(),
		})
		for _, ch := range s.Channels.all() {
			if ch.secret() {
				continue
			}

			c.sendRPL(s.name, rplList{
				client:  c.nickname(),
				channel: ch.name(),
				count:   ch.count(),
				topic:   ch.topic().text,
			})
		}
		c.sendRPL(s.name, rplListEnd{
			client: c.nickname(),
		})
		return
	}
}
