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
	server.gauges["ircd_clients"].Inc()

	// starts goroutines for procesing incoming and outgoing messages
	go HandleConnectionOut(client, server)
	go HandleConnectionIn(client, server)

	// read input from client

	scanner := bufio.NewScanner(client.connection)
	for scanner.Scan() {
		line := scanner.Text()
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", client.connection.RemoteAddr())
			break
		}

		line = strings.Trim(line, "\r\n")

		// PING
		// https://modern.ircdocs.horse/#ping-message
		// https://modern.ircdocs.horse/#pong-message
		if strings.HasPrefix(line, "PING") {
			client.send <- strings.Replace(line, "PING", "PONG", 1)
			continue
		}

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
