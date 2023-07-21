package ircd

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

func HandleConnectionOut(client *Client, server *Server) {
	defer func() {
		client.stop <- true
	}()

	writer := bufio.NewWriter(client.writer)

	for message := range client.send {
		log.Debug().Msgf("%s: %s", client.Prefix(), message)
		_, err := writer.WriteString(message + "\r\n")
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.ip)
			break
		}
		err = writer.Flush()
		if err != nil {
			log.Error().Err(err).Msg("unable to flush writer")
			break
		}
	}
}
