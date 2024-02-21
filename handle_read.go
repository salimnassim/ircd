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
	c, err := newClient(conn, id)
	if err != nil {
		conn.Close()
		log.Error().Err(err).Msg("unable to create client")
		return
	}

	// add client to store
	s.clients.add(clientID(id), c)
	metrics.Clients.Inc()

	// starts goroutines for procesing incoming and outgoing messages
	go handleConnectionIn(c, s)
	go handleConnectionOut(c, s)

	// read input from client
	scanner := bufio.NewScanner(c.reader)
	for scanner.Scan() {
		line := scanner.Text()
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", c.ip)
			break
		}

		line = strings.Trim(line, "\r\n")

		// PING
		// https://modern.ircdocs.horse/#ping-message
		// https://modern.ircdocs.horse/#pong-message
		if strings.HasPrefix(line, "PING") {
			c.setPing(time.Now().Unix())
			c.send <- strings.Replace(line, "PING", "PONG", 1)
			metrics.Pings.Inc()
			continue
		}

		// send to receive channel
		c.recv <- line
	}

	// cleanup
	err = conn.Close()
	if err != nil {
		log.Error().Err(err).Msg("unable to close client on write handler")
	}
	err = s.removeClient(c)
	if err != nil {
		log.Error().Err(err).Msg("unable to remove client on write handler")
	}
}
