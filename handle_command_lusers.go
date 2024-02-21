package ircd

import "github.com/salimnassim/ircd/metrics"

func handleLusers(s *server, c *client, m Message) {
	clients, channels := s.stats()

	c.sendRPL(s.name, rplLuserClient{
		client:    c.nickname(),
		users:     clients,
		invisible: 0,
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

	metrics.Lusers.Inc()
}
