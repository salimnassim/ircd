package ircd

import (
	"fmt"
	"time"
)

func handleConnectionPong(c *client, s *server) {
	pingDuration := time.Duration(s.pingFrequency) * time.Second
	pongDuration := time.Duration(s.pongMaxLatency) * time.Second

	var timer <-chan time.Time
	alive := true
	for alive {
		select {
		case <-c.killPong:
			alive = false
		case <-c.ponged:
			timer = nil
		case <-time.After(pingDuration):
			c.sendCommand(pingCommand{
				text: s.name,
			})
			timer = time.After(pongDuration)
		case <-timer:
			c.kill(fmt.Sprintf("Timeout after %d seconds", s.pongMaxLatency))
			continue
		}
	}
}
