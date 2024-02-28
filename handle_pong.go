package ircd

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

func handleConnectionPong(c *client, s *server) {
	pingDuration := time.Duration(s.pingFrequency) * time.Second
	pongDuration := time.Duration(s.pongMaxLatency) * time.Second

	var timer <-chan time.Time
	alive := true
	for alive {
		select {
		case <-c.killPong:
			log.Debug().Str("nick", c.nickname()).Msg("Killed pong")
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
		}
	}
}
