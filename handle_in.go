package ircd

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionIn(client *Client, server *Server) {
	defer func() {
		client.stop <- true
	}()

	for message := range client.recv {
		parsed, err := ParseMessage(message)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse message in handler: %s", message)
			continue
		}

		log.Debug().Msgf("%s", parsed.Raw)

		// NICK
		// https://modern.ircdocs.horse/#nick-message
		if parsed.Command == "NICK" {
			handleNick(server, client, parsed)
			continue
		}

		// USER
		// https://modern.ircdocs.horse/#user-message
		if parsed.Command == "USER" {
			handleUser(server, client, parsed)
			continue
		}

		// LUSERS
		// https://modern.ircdocs.horse/#lusers-message
		if parsed.Command == "LUSERS" {
			handleLusers(server, client, parsed)
			continue
		}

		// JOIN
		// https://modern.ircdocs.horse/#join-message
		if parsed.Command == "JOIN" {
			handleJoin(server, client, parsed)
			continue
		}

		// PART
		// https://modern.ircdocs.horse/#part-message
		if parsed.Command == "PART" {
			handlePart(server, client, parsed)
			continue
		}

		// TOPIC
		// https://modern.ircdocs.horse/#topic-message
		if parsed.Command == "TOPIC" {
			handleTopic(server, client, parsed)
			continue
		}

		// PRIVMSG
		// https://modern.ircdocs.horse/#privmsg-message
		if parsed.Command == "PRIVMSG" {
			handlePrivmsg(server, client, parsed)
			continue
		}

		// WHOIS
		// https://modern.ircdocs.horse/#whois-message
		// if parsed.Command == "WHOIS" {
		// 	if !client.handshake {
		// 		client.send <- fmt.Sprintf(":%s 451 :You have not registered.", server.name)
		// 		continue
		// 	}

		// 	target := parsed.Params[0]
		// 	whois, exists := server.clients.Whois(target, server.channels)
		// 	if !exists {
		// 		client.send <- fmt.Sprintf(
		// 			":%s 401 %s :no such nick",
		// 			server.name,
		// 			client.nickname)
		// 		continue
		// 	}

		// 	// https://modern.ircdocs.horse/#rplwhoisuser-311
		// 	client.send <- fmt.Sprintf(
		// 		":%s 311 %s %s %s %s * :%s",
		// 		server.name,
		// 		client.nickname,
		// 		whois.nickname,
		// 		whois.username,
		// 		whois.hostname,
		// 		whois.realname,
		// 	)

		// 	// https://modern.ircdocs.horse/#rplwhoischannels-319
		// 	client.send <- fmt.Sprintf(
		// 		":%s 319 %s %s :%s",
		// 		server.name,
		// 		client.nickname,
		// 		whois.nickname,
		// 		strings.Join(whois.channels, " "),
		// 	)

		// 	continue
		// }

		// https://modern.ircdocs.horse/#who-message
		if parsed.Command == "WHO" {
			if !client.handshake {
				client.send <- fmt.Sprintf(":%s 451 :You have not registered.", server.name)
				continue
			}

			if len(parsed.Params) == 0 {
				client.send <- fmt.Sprintf(":%s 461 %s WHO :Not enough parameters",
					server.name,
					client.nickname)
				continue
			}

			target := parsed.Params[0]
			if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
				channel, ok := server.channels.Get(target)
				if !ok {
					client.send <- fmt.Sprintf(":%s 403 %s :No such channel.", server.name, target)
					continue
				}

				who := channel.Who()
				for _, v := range who {
					client.send <- fmt.Sprintf(":%s 352 %s %s", server.name, client.Nickname(), v)
				}
				client.send <- fmt.Sprintf(":%s 315 %s :End of WHO list", server.name, target)
				continue
			}

		}

	}

}
