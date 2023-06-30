package ircd

import "github.com/rs/zerolog/log"

func HandleConnectionOut(client *Client) {
	for message := range client.out {
		n, err := client.Write(message + "\r\n")
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.connection.RemoteAddr())
			break
		}
		log.Info().Msgf("out(%5d)> %s", n, message)
	}
}
