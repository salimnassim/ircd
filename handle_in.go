package ircd

import (
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionIn(client *Client) {
	for message := range client.In {
		parsed, err := Parse(message)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse message in handler: %s", message)
		}
		log.Info().Msgf(" in(%5d)> %s", len(message), message)
		log.Debug().Msgf("prefix: %s, command: %s, args: %s", parsed.Prefix, parsed.Command, strings.Join(parsed.Params, "|"))
	}
}
