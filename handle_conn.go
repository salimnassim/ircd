package ircd

import (
	"net"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd/metrics"
)

func handleConnection(conn net.Conn, s *server) {
	id := uuid.Must(uuid.NewRandom()).String()
	c, err := newClient(conn, id)
	if err != nil {
		log.Error().Err(err).Msg("unable to create client")
		return
	}

	s.Clients.add(c)
	metrics.Clients.Inc()

	// starts goroutines for procesing incoming and outgoing messages
	go handleConnectionIn(c, s)
	go handleConnectionOut(c, s)
	handleConnectionPong(c, s)

	s.cleanup(c)
	metrics.Clients.Dec()
}
