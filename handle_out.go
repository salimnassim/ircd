package ircd

import (
	"github.com/rs/zerolog/log"
)

func HandleConnectionOut(client *Client, server *Server) {
	defer func() {
		client.stop <- true
	}()

	for message := range client.send {
		log.Debug().Msgf("%s: %s", client.Prefix(), message)
		_, err := client.Write(message + "\r\n")
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.ip)
			break
		}
	}
}
