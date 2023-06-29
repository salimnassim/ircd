package ircd

import "github.com/rs/zerolog/log"

func HandleConnectionOut(client *Client) {
	for message := range client.Out {
		n, err := client.Connection.Write([]byte(message + "\r\n"))
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.Connection.RemoteAddr())
			break
		}
		log.Info().Msgf("out(%5d)> %s", n, message)
	}
}
