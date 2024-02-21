package ircd

import (
	"bufio"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd/metrics"
)

func HandleConnection(conn net.Conn, s *server) {
	id := uuid.Must(uuid.NewRandom()).String()
	client, err := NewClient(conn, id)
	if err != nil {
		conn.Close()
		log.Error().Err(err).Msg("unable to create client")
		return
	}

	// add client to store
	s.clients.Add(clientID(id), client)
	metrics.Clients.Inc()

	// starts goroutines for procesing incoming and outgoing messages
	go handleConnectionIn(client, s)
	go handleConnectionOut(client, s)

	// read input from client
	scanner := bufio.NewScanner(client.reader)
	for scanner.Scan() {
		line := scanner.Text()
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", client.ip)
			break
		}

		line = strings.Trim(line, "\r\n")

		// PING
		// https://modern.ircdocs.horse/#ping-message
		// https://modern.ircdocs.horse/#pong-message
		if strings.HasPrefix(line, "PING") {
			client.setPing(time.Now().Unix())
			client.send <- strings.Replace(line, "PING", "PONG", 1)
			metrics.Pings.Inc()
			continue
		}

		// send to receive channel
		client.recv <- line
	}

	// cleanup
	err = conn.Close()
	if err != nil {
		log.Error().Err(err).Msg("unable to close client on write handler")
	}
	err = s.removeClient(client)
	if err != nil {
		log.Error().Err(err).Msg("unable to remove client on write handler")
	}
}
