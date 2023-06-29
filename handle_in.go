package ircd

import "github.com/rs/zerolog/log"

func HandleConnectionIn(client *Client) {
	for message := range client.In {
		log.Info().Msgf(" in(%5d)> %s", len(message), message)
	}
}
