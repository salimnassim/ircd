package ircd

import (
	"github.com/rs/zerolog/log"
)

func handleConnectionOut(c *client, s *server) {
	defer func() {
		c.stop <- "broken pipe"
	}()

	for message := range c.send {
		log.Debug().Str("nick", c.nickname()).Msgf("%s", message)
		_, err := c.write(message + "\r\n")
		if err != nil {
			log.Error().Err(err).Msgf("cant write to client '%s'", c.id)
			break
		}
	}
}
