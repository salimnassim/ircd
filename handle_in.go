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
			if !client.handshake {
				client.send <- fmt.Sprintf(":%s 451 :You have not registered.",
					server.name)
				continue
			}

			targets := strings.Split(parsed.Params[0], ",")

			for _, target := range targets {
				// try to get channel
				channel, exists := server.channels.GetByName(target)
				if !exists {
					client.send <- fmt.Sprintf(":%s 403 %s :No such channel.",
						server.name, target)
					continue
				}

				// remove client
				channel.RemoveClient(client)

				// broadcast that user has left the channel
				channel.Broadcast(fmt.Sprintf(":%s PART %s",
					client.Prefix(), target), client.id, false)
			}

			// todo: check if last user

			continue
		}

		// TOPIC
		// https://modern.ircdocs.horse/#topic-message
		if parsed.Command == "TOPIC" {
			if !client.handshake {
				client.send <- fmt.Sprintf(":%s 451 :You have not registered.", server.name)
				continue
			}

			target := parsed.Params[0]
			remainder := strings.Join(parsed.Params[1:len(parsed.Params)], " ")

			// channel name max length is 50, check for allowed channel prefixes
			if len(target) > 50 || !(strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&")) {
				client.send <- fmt.Sprintf(":%s 403 %s :No such channel (2).",
					server.name, target)
				continue
			}

			// try to get channel
			channel, exists := server.channels.GetByName(target)
			if !exists {
				client.send <- fmt.Sprintf(":%s 403 %s :No such channel.", server.name, target)
				continue
			}

			// set topic
			channel.SetTopic(remainder, client.nickname)

			// get topic
			topic := channel.Topic()

			// broadcast new topic to clients on channel
			channel.Broadcast(
				fmt.Sprintf(":%s 332 %s %s :%s",
					server.name,
					client.nickname,
					channel.name,
					topic.text),
				client.id, false)

			continue
		}

		// PRIVMSG
		// https://modern.ircdocs.horse/#privmsg-message
		if parsed.Command == "PRIVMSG" {
			if !client.handshake {
				client.send <- fmt.Sprintf(":%s 451 :You have not registered.", server.name)
				continue
			}

			targets := strings.Split(parsed.Params[0], ",")
			message := strings.Join(parsed.Params[1:len(parsed.Params)], " ")

			for _, target := range targets {
				// is channel
				if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
					channel, exists := server.channels.GetByName(target)
					if !exists {
						client.send <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
							server.name,
							client.nickname)
						continue
					}

					// is user a member of the channel?
					if !server.channels.IsMember(client, channel) {
						client.send <- fmt.Sprintf(":%s 442 %s :You're not on that channel",
							server.name,
							client.nickname)
						continue
					}

					// send message to channel
					channel.Broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s",
						client.Prefix(), channel.name, message), client.id, true)
					promPrivmsgChannel.Inc()
					continue
				}
				// is user
				dest, exists := server.clients.GetByNickname(target)
				if !exists {
					client.send <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.name,
						client.nickname)
					continue
				}
				dest.send <- fmt.Sprintf(":%s PRIVMSG %s :%s",
					client.nickname, dest.nickname, message)
				promPrivmsgClient.Inc()
				continue
			}

		}

		// WHOIS
		// https://modern.ircdocs.horse/#whois-message
		if parsed.Command == "WHOIS" {
			if !client.handshake {
				client.send <- fmt.Sprintf(":%s 451 :You have not registered.", server.name)
				continue
			}

			target := parsed.Params[0]
			whois, exists := server.clients.Whois(target, server.channels)
			if !exists {
				client.send <- fmt.Sprintf(
					":%s 401 %s :no such nick",
					server.name,
					client.nickname)
				continue
			}

			// https://modern.ircdocs.horse/#rplwhoisuser-311
			client.send <- fmt.Sprintf(
				":%s 311 %s %s %s %s * :%s",
				server.name,
				client.nickname,
				whois.nickname,
				whois.username,
				whois.hostname,
				whois.realname,
			)

			// https://modern.ircdocs.horse/#rplwhoischannels-319
			client.send <- fmt.Sprintf(
				":%s 319 %s %s :%s",
				server.name,
				client.nickname,
				whois.nickname,
				strings.Join(whois.channels, " "),
			)

			continue
		}

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
				channel, ok := server.channels.GetByName(target)
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
