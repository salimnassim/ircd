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
	if err != nil {
		log.Error().Err(err).Msg("unable to create client")
		return
	}

	defer conn.Close()
	defer s.removeClient(c)
	defer metrics.Clients.Dec()

	s.Clients.add(c)
	metrics.Clients.Inc()

	// starts goroutines for procesing incoming and outgoing messages
	go handleConnectionIn(c, s)
	go handleConnectionOut(c, s)

	// read input from client
	go func() {
		defer func() {
			c.stop <- "Broken pipe."
		}()

		reader := bufio.NewReader(conn)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			if scanner.Err() != nil {
				c.stop <- err.Error()
				break
			}
			line := scanner.Text()
			if err != nil {
				c.stop <- err.Error()
				break
			}
			line = strings.Trim(line, "\r\n")
			c.recv(line)
		}
	}()

	pingDuration := time.Duration(s.pingFrequency) * time.Second
	pongDuration := time.Duration(s.pongMaxLatency) * time.Second

	var timer <-chan time.Time
	for c.alive {
		select {
		case e := <-c.stop:
			if e != "quit" {
				for _, ch := range s.Channels.memberOf(c) {
					ch.broadcastCommand(partCommand{
						prefix:  c.prefix(),
						channel: ch.name(),
						text:    fmt.Sprintf("Quit: %s", e),
					}, c.clientID, true)
				}
			}
			c.alive = false
		case <-c.ponged:
			timer = nil
		case <-timer:
			for _, ch := range s.Channels.memberOf(c) {
				ch.broadcastCommand(partCommand{
					prefix:  c.prefix(),
					channel: ch.name(),
					text:    fmt.Sprintf("Quit: Timeout after %d seconds", s.pongMaxLatency),
				}, c.clientID, true)
			}
			c.alive = false
		case <-time.After(pingDuration):
			c.sendCommand(pingCommand{
				text: s.name,
			})
			timer = time.After(pongDuration)
		}
	}
}
