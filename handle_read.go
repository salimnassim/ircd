package ircd

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionRead(client *Client, server *Server) {
	reader := bufio.NewReader(client.connection)

	validNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\|]{2,16})`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile nickname validation regex")
	}

	for {

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", client.connection.RemoteAddr())
			break
		}
		line = strings.Trim(line, "\r\n")

		client.In <- line
		split := strings.Split(line, " ")

		if strings.HasPrefix(line, "PING") {
			pong := strings.Replace(line, "PING", "PONG", 1)
			client.Out <- pong
			continue
		}
		if strings.HasPrefix(line, "NICK") {
			nickname := split[1]

			// todo: validate nickname
			ok := validNick.MatchString(nickname)
			if !ok {
				client.Out <- fmt.Sprintf(":%s 432 * %s :Erroneus nickname.", server.Name, nickname)
				continue
			}
			if err != nil {
				log.Error().Err(err).Msg("unable to match nickname regex")
				continue
			}

			_, exists := server.ClientByNickname(split[1])
			if exists {
				client.Out <- fmt.Sprintf(":%s 433 * %s :Nickname is already in use.", server.Name, nickname)
				continue
			}

			client.Nickname = nickname

			if !client.Handshake {
				client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Your ID is: %s", client.Nickname, client.ID)
				client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Looking up your hostname..", client.Nickname)
				addr, err := net.LookupAddr(strings.Split(client.connection.RemoteAddr().String(), ":")[0])
				if err != nil {
					client.SetHostname(strings.Split(client.connection.RemoteAddr().String(), ":")[0])
				} else {
					client.SetHostname(addr[0])
				}

				client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Checking ident (not really)", client.Nickname)

				client.Out <- fmt.Sprintf(":%s 001 %s :Welcome to the IRC network!", server.Name, client.Nickname)
				client.Out <- fmt.Sprintf(":%s 376 %s :End of /MOTD command", server.Name, client.Nickname)
				client.Handshake = true
			}
			continue
		}

		if strings.HasPrefix(line, "PRIVMSG") {
			target := split[1]
			message := strings.Split(line, ":")[1]

			// to channel
			if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "?") || strings.HasPrefix(target, "~") {
				channel, err := server.Channel(target)
				if err != nil {
					client.Out <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.Name,
						client.Nickname)
					log.Error().Err(err).Msgf("privmsg channel %s does not exist", target)
					continue
				}
				channel.Broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s", client.Target(), target, message), client, true)
				server.Counters["ircd_channels_privmsg"].Inc()
			} else {
				// to user
				tc, found := server.ClientByNickname(target)
				if !found {
					client.Out <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.Name,
						client.Nickname)
					log.Error().Err(err).Msgf("privmsg user %s does not exist", target)
					continue
				}
				tc.Out <- fmt.Sprintf(":%s PRIVMSG %s :%s", client.Nickname, tc.Nickname, message)
				server.Counters["ircd_clients_privmsg"].Inc()
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
				if topic.Topic == "" {
					client.Out <- fmt.Sprintf(":%s 331 %s %s :no topic set",
						server.Name,
						client.Nickname,
						channel.Name,
					)
				} else {
					client.Out <- fmt.Sprintf(":%s 332 %s %s :%s",
						server.Name,
						client.Nickname,
						channel.Name,
						topic.Topic,
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

		if strings.HasPrefix(strings.ToUpper(line), "LUSERS") {
			client.Out <- fmt.Sprintf(":%s 251 %s :There are %d users and %d invisible on %d servers",
				server.Name,
				client.Nickname,
				-1,
				-1,
				-1,
			)
			client.Out <- fmt.Sprintf(":%s 254 %s %d :channels formed",
				server.Name,
				client.Nickname,
				-1)
			continue
		}

		if strings.HasPrefix(line, "TOPIC") {
			target := split[1]
			message := strings.Replace(split[2], ":", "", 1)

			channel, err := server.Channel(target)
			if err != nil {
				log.Error().Err(err).Msgf("tried to change topic %s", target)
				continue
			}

			channel.SetTopic(message, client.Nickname)
			channel.Broadcast(
				fmt.Sprintf(":%s 332 %s %s :%s",
					server.Name,
					client.Nickname,
					channel.Name,
					message),
				client, false)
			continue
		}

		if strings.HasPrefix(line, "QUIT") {
			break
		}

	}

	log.Info().Msgf("closing client from connection read (%s)", client.connection.RemoteAddr())
	err = client.Close()
	if err != nil {
		log.Error().Err(err).Msg("unable to close client on read handler")
	}
	err = server.RemoveClient(client)
	if err != nil {
		log.Error().Err(err).Msg("unable to remove client on read handler")
	}
}
