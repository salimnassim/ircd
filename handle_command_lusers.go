package ircd

import "github.com/salimnassim/ircd/metrics"

func handleLusers(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	visible, invisible, channels := s.stats()

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

	metrics.Lusers.Inc()
}
