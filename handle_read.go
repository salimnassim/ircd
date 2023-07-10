package ircd

import (
	"bufio"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func HandleConnectionRead(connection net.Conn, server *Server) {
	id := uuid.Must(uuid.NewRandom()).String()
	client, err := NewClient(connection, id)
	if err != nil {
		connection.Close()
		log.Error().Err(err).Msg("unable to create client")
		return
	}

	// add client to store
	server.clients.Add(client)

	// starts goroutines for procesing incoming and outgoing messages
	go HandleConnectionOut(client, server)
	go HandleConnectionIn(client, server)

	// read input from client
	reader := bufio.NewReader(client.connection)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", client.connection.RemoteAddr())
			break
		}
		line = strings.Trim(line, "\r\n")
		// send to receive channel
		client.recv <- line
	}

	// cleanup
	err = client.Close()
	if err != nil {
		log.Error().Err(err).Msg("unable to close client on write handler")
	}
	err = server.RemoveClient(client)
	if err != nil {
		log.Error().Err(err).Msg("unable to remove client on write handler")
	}
}
