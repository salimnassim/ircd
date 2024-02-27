package ircd

import (
	"github.com/rs/zerolog/log"
)

func handleConnectionOut(c *client, s *server) {
	for message := range c.out {
		log.Debug().Str("nick", c.nickname()).Msgf("%s", message)
		_, err := c.conn.Write([]byte(message + "\r\n"))
		if err != nil {
			log.Error().Err(err).Msgf("cant write to client '%s'", c.clientID)
			break
		}
	}

	if !c.alive {
		c.kill("Broken pipe")
	}
}
