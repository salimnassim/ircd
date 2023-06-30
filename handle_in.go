package ircd

import (
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionIn(client *Client, server *Server) {
	for message := range client.in {
		parsed, err := Parse(message)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse message in handler: %s", message)
		}

		// PING
		if parsed.Command == "PING" {
			pong := strings.Replace(parsed.Raw, "PING", "PONG", 1)
			client.out <- pong
			continue
		}

		// NICK
		if parsed.Command == "NICK" {
			// has enough params?
			if len(parsed.Params) != 1 {
				client.out <- fmt.Sprintf(":%s 461 * %s :Not enough parameters.", server.name, parsed.Command)
				log.Error().Msg("not enough params for nick")
				continue
			}

			// validate nick
			ok := server.regex["nick"].MatchString(parsed.Params[0])
			if !ok {
				client.out <- fmt.Sprintf(":%s 432 * %s :Erroneus nickname.", server.name, parsed.Params[0])
				continue
			}

			_, exists := server.ClientByNickname(parsed.Params[0])
			if exists {
				client.out <- fmt.Sprintf(":%s 433 * %s :Nickname is already in use.", server.name, parsed.Params[0])
				continue
			}

			// set nickname
			client.nickname = parsed.Params[0]

			// check for handshake
			if !client.handshake {
				client.out <- fmt.Sprintf("NOTICE %s :AUTH :*** Your ID is: %s", client.nickname, client.id)
				client.out <- fmt.Sprintf("NOTICE %s :AUTH :*** Looking up your hostname..", client.nickname)

				// lookup address
				// todo: resolver
				addr, err := net.LookupAddr(strings.Split(client.connection.RemoteAddr().String(), ":")[0])
				if err != nil {
					// if it cant be resolved use ip
					client.SetHostname(strings.Split(client.connection.RemoteAddr().String(), ":")[0])
				} else {
					client.SetHostname(addr[0])
				}

				client.out <- fmt.Sprintf(":%s 001 %s :Welcome to the IRC network! 🎂", server.name, client.nickname)
				client.out <- fmt.Sprintf(":%s 376 %s :End of /MOTD command", server.name, client.nickname)
				client.handshake = true

				server.AddClient(client)
			}

			log.Debug().Msgf("prefix: %s, command: %s, args: %s", parsed.Prefix, parsed.Command, strings.Join(parsed.Params, "|"))
			continue
		}

		// USER
		if parsed.Command == "USER" {
			if len(parsed.Params) < 4 {
				client.out <- fmt.Sprintf(":%s 461 %s %s :Not enough parameters.", server.name, client.nickname, strings.Join(parsed.Params, " "))
				continue
			}

			if client.username != "" {
				client.out <- fmt.Sprintf(":%s 462 %s :You may not reregister.", server.name, client.nickname)
				continue
			}

			// todo: validate
			username := parsed.Params[0]
			realname := parsed.Params[3]

			client.username = username
			client.realname = realname
			continue
		}

		// LUSERS
		if parsed.Command == "LUSERS" {
			clients, channels := server.Stats()
			client.out <- fmt.Sprintf(":%s 461 %s :There are %d active users on %d channels.", server.name, client.nickname, clients, channels)
			continue
		}

		log.Info().Msgf(" in(%5d)> %s", len(message), message)
	}
}
