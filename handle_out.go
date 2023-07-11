package ircd

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

func HandleConnectionOut(client *Client, server *Server) {
	writer := bufio.NewWriter(client.connection)

	for message := range client.send {
		log.Debug().Msgf("%s: %s", client.Prefix(), message)

		_, err := writer.WriteString(message + "\r\n")
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.connection.RemoteAddr())
			break
		}
		writer.Flush()
	}
}
