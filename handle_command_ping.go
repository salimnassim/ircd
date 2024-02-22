package ircd

import (
	"strings"
	"time"

	"github.com/salimnassim/ircd/metrics"
)

func handlePing(s *server, c *client, m message) {
	c.setPing(time.Now().Unix())
	c.send <- strings.Replace(m.raw, "PING", "PONG", 1)
	metrics.Ping.Inc()
}
