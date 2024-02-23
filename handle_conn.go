package ircd

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd/metrics"
)

func handleConnection(conn net.Conn, s *server) {
	id := uuid.Must(uuid.NewRandom()).String()
	c, err := newClient(conn, id)

	defer conn.Close()
	defer s.removeClient(c)

	if err != nil {
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
	go func() {
		scanner := bufio.NewScanner(c.reader)
		for scanner.Scan() {
			if scanner.Err() != nil {
				c.stop <- true
				break
			}
			line := scanner.Text()
			if err != nil {
				c.stop <- true
				break
			}
			line = strings.Trim(line, "\r\n")
			c.recv <- line
		}
	}()

	pingDuration := time.Duration(s.pingFrequency) * time.Second
	pongDuration := time.Duration(s.pongMaxLatency) * time.Second

	var timer <-chan time.Time
	for {
		select {
		case <-c.stop:
			log.Debug().Msg("SELECT STOP")
			return

		case <-c.pong:
			timer = nil
		case <-timer:
			for _, ch := range s.channels.memberOf(c) {
				ch.broadcast(
					fmt.Sprintf(":%s PART %s :Quit: Timeout after %d seconds.", c.prefix(), ch.name, s.pongMaxLatency),
					c.id,
					false,
				)
			}
			return
		case <-time.After(pingDuration):
			c.send <- fmt.Sprintf("PING %s", s.name)
			timer = time.After(pongDuration)
		}
	}
}
