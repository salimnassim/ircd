package ircd

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func HandleConnectionRead(connection net.Conn, server *Server) {
	log.Info().Msgf("handling connection")

	id := uuid.Must(uuid.NewRandom()).String()
	client, err := NewClient(connection, id)
	if err != nil {
		connection.Close()
		log.Error().Err(err).Msg("unable to create client")
		return
	}

	go HandleConnectionIn(client, server)
	go HandleConnectionOut(client, server)

	reader := bufio.NewReader(client.connection)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", client.connection.RemoteAddr())
			break
		}
		line = strings.Trim(line, "\r\n")

		// send line to recv
		client.recv <- line

		// split
		split := strings.Split(line, " ")

		if strings.HasPrefix(line, "PRIVMSG") {
			target := split[1]
			message := strings.Split(line, ":")[1]

			// to channel
			if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "?") || strings.HasPrefix(target, "~") {
				channel, ok := server.Channel(target)
				if !ok {
					client.send <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.name,
						client.nickname)
					log.Error().Err(err).Msgf("privmsg channel %s does not exist", target)
					continue
				}
				channel.Broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s", client.Prefix(), target, message), client, true)
				server.counters["ircd_channels_privmsg"].Inc()
			} else {
				// to user
				tc, found := server.ClientByNickname(target)
				if !found {
					client.send <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.name,
						client.nickname)
					log.Error().Err(err).Msgf("privmsg user %s does not exist", target)
					continue
				}
				tc.send <- fmt.Sprintf(":%s PRIVMSG %s :%s", client.nickname, tc.nickname, message)
				server.counters["ircd_clients_privmsg"].Inc()
			}
			continue
		}

		if strings.HasPrefix(line, "QUIT") {
			break
		}

	}

	log.Info().Msgf("closing client from connection read (%s)", client.connection.RemoteAddr())

	err = client.Close()
	if err != nil {
		log.Error().Err(err).Msg("unable to close client on write handler")
	}
	err = server.RemoveClient(client)
	if err != nil {
		log.Error().Err(err).Msg("unable to remove client on write handler")
	}
}
