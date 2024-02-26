package ircd

import (
	"github.com/rs/zerolog/log"
)

func handleConnectionIn(c *client, s *server) {
	defer c.kill("Broken pipe.")

	for message := range c.in {
		parsed, err := parseMessage(message)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse message in handler: %s", message)
			continue
		}

		log.Debug().Str("nick", c.nickname()).Msgf("%s", parsed.raw)

		s.router.handle(s, c, parsed)
	}
}
