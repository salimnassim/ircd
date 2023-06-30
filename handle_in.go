package ircd

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionIn(client *Client, server *Server) {

	rgxNickname, err := regexp.Compile(`([a-zA-Z0-9\[\]\|]{2,16})`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile nickname validation regex")
	}

	for message := range client.In {
		parsed, err := Parse(message)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse message in handler: %s", message)
		}

		// if parsed.Command == "PING" {
		// 	pong := strings.Replace(parsed.Raw, "PING", "PONG", 1)
		// 	client.Out <- pong
		// 	continue
		// }

		// NICK
		if parsed.Command == "NICK" {

			// has enough params?
			if len(parsed.Params) != 1 {
				client.Out <- fmt.Sprintf(":%s 461 * %s :Not enough parameters.", server.Name, parsed.Command)
				log.Error().Msg("not enough params for nick")
				continue
			}

			// validate nick
			ok := rgxNickname.MatchString(parsed.Params[0])
			if !ok {
				client.Out <- fmt.Sprintf(":%s 432 * %s :Erroneus nickname.", server.Name, parsed.Params[0])
				continue
			}

			_, exists := server.ClientByNickname(parsed.Params[0])
			if exists {
				client.Out <- fmt.Sprintf(":%s 433 * %s :Nickname is already in use.", server.Name, parsed.Params[0])
				continue
			}

			client.Nickname = parsed.Params[0]

			// check for handshake
			if !client.Handshake {
				client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Your ID is: %s", client.Nickname, client.ID)
				client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Looking up your hostname..", client.Nickname)

				// lookup address
				addr, err := net.LookupAddr(strings.Split(client.connection.RemoteAddr().String(), ":")[0])
				if err != nil {
					// if it cant be resolved use ip
					client.SetHostname(strings.Split(client.connection.RemoteAddr().String(), ":")[0])
				} else {
					client.SetHostname(addr[0])
				}

				// client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Checking ident", client.Nickname)

				client.Out <- fmt.Sprintf(":%s 001 %s :Welcome to the IRC network! 🎂", server.Name, client.Nickname)
				client.Out <- fmt.Sprintf(":%s 376 %s :End of /MOTD command", server.Name, client.Nickname)
				client.Handshake = true
			}

			log.Debug().Msgf("prefix: %s, command: %s, args: %s", parsed.Prefix, parsed.Command, strings.Join(parsed.Params, "|"))
			continue
		}

		// USER
		if parsed.Command == "USER" {
			if len(parsed.Params) < 4 {
				client.Out <- fmt.Sprintf(":%s 461 %s %s :Not enough parameters.", server.Name, client.Nickname, strings.Join(parsed.Params, " "))
				continue
			}

			if client.Username != "" {
				client.Out <- fmt.Sprintf(":%s 462 %s :You may not reregister.", server.Name, client.Nickname)
				continue
			}

			username := parsed.Params[0]
			realname := parsed.Params[3]

			client.Username = username
			client.Realname = realname
			continue
		}

		log.Info().Msgf(" in(%5d)> %s", len(message), message)
	}
}
