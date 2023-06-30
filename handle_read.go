package ircd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionRead(client *Client, server *Server) {
	reader := bufio.NewReader(client.connection)

	for {

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", client.connection.RemoteAddr())
			break
		}
		line = strings.Trim(line, "\r\n")

		client.in <- line
		split := strings.Split(line, " ")

		if strings.HasPrefix(line, "PRIVMSG") {
			target := split[1]
			message := strings.Split(line, ":")[1]

			// to channel
			if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "?") || strings.HasPrefix(target, "~") {
				channel, err := server.Channel(target)
				if err != nil {
					client.out <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.name,
						client.nickname)
					log.Error().Err(err).Msgf("privmsg channel %s does not exist", target)
					continue
				}
				channel.Broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s", client.Target(), target, message), client, true)
				server.counters["ircd_channels_privmsg"].Inc()
			} else {
				// to user
				tc, found := server.ClientByNickname(target)
				if !found {
					client.out <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.name,
						client.nickname)
					log.Error().Err(err).Msgf("privmsg user %s does not exist", target)
					continue
				}
				tc.out <- fmt.Sprintf(":%s PRIVMSG %s :%s", client.nickname, tc.nickname, message)
				server.counters["ircd_clients_privmsg"].Inc()
			}
			continue
		}

		if strings.HasPrefix(line, "JOIN") {
			targets := strings.Split(split[1], ",")

			for _, target := range targets {
				channel, err := server.Channel(target)
				if err != nil && channel == nil {
					channel = server.CreateChannel(target)
				}
				err = channel.AddClient(client, "")
				if err != nil {
					log.Error().Err(err).Msgf("cant join channel %s", target)
					continue
				}

				channel.Broadcast(
					fmt.Sprintf(":%s JOIN %s",
						client.Target(), target),
					client,
					false,
				)

				topic := channel.Topic()
				if topic.text == "" {
					client.out <- fmt.Sprintf(":%s 331 %s %s :no topic set",
						server.name,
						client.nickname,
						channel.name,
					)
				} else {
					client.out <- fmt.Sprintf(":%s 332 %s %s :%s",
						server.name,
						client.nickname,
						channel.name,
						topic.text,
					)
				}
			}
			continue
		}

		if strings.HasPrefix(line, "PART") {
			target := split[1]

			channel, err := server.Channel(target)
			if err != nil {
				log.Error().Err(err).Msgf("tried to part channel that does not exist %s", target)
			}

			channel.Broadcast(fmt.Sprintf(":%s PART %s", client.Target(), target), client, false)
			channel.RemoveClient(client)
		}

		if strings.HasPrefix(line, "TOPIC") {
			target := split[1]
			message := strings.Replace(split[2], ":", "", 1)

			channel, err := server.Channel(target)
			if err != nil {
				log.Error().Err(err).Msgf("tried to change topic %s", target)
				continue
			}

			channel.SetTopic(message, client.nickname)
			channel.Broadcast(
				fmt.Sprintf(":%s 332 %s %s :%s",
					server.name,
					client.nickname,
					channel.name,
					message),
				client, false)
			continue
		}

		if strings.HasPrefix(line, "QUIT") {
			break
		}

	}

	log.Info().Msgf("closing client from connection read (%s)", client.connection.RemoteAddr())
	err := client.Close()
	if err != nil {
		log.Error().Err(err).Msg("unable to close client on read handler")
	}
	err = server.RemoveClient(client)
	if err != nil {
		log.Error().Err(err).Msg("unable to remove client on read handler")
	}
}
