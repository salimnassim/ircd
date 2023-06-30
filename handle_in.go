package ircd

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionIn(client *Client, server *Server) {
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
			log.Debug().Msgf("prefix: %s, command: %s, args: %s", parsed.Prefix, parsed.Command, strings.Join(parsed.Params, "|"))
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
