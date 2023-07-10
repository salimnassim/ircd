package ircd

import "github.com/rs/zerolog/log"

func HandleConnectionOut(client *Client, server *Server) {
	for message := range client.send {
		_, err := client.Write(message + "\r\n")
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.connection.RemoteAddr())
			break
		}

		log.Debug().Msgf("%s: %s", client.Prefix(), message)
	}
}
