package ircd

import (
	"github.com/rs/zerolog/log"
)

func handleConnectionOut(c *client, s *server) {
	defer func() {
		c.stop <- true
	}()

	for message := range c.send {
		log.Debug().Str("nick", c.nickname()).Msgf("%s", message)
		_, err := c.write(message + "\r\n")
		if err != nil {
			break
		}
	}
}
