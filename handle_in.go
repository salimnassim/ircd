package ircd

import (
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleConnectionIn(client *Client, server *Server) {
	for message := range client.recv {
		parsed, err := Parse(message)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse message in handler: %s", message)
		}

		log.Debug().Msgf(" in(%5d)> [command: %5s, prefix: %s, params: %s] %s",
			len(message), parsed.Command, parsed.Prefix, strings.Join(parsed.Params, ","), parsed.Raw)

		// PING
		// https://modern.ircdocs.horse/#ping-message
		// https://modern.ircdocs.horse/#pong-message
		if parsed.Command == "PING" {
			pong := strings.Replace(parsed.Raw, "PING", "PONG", 1)
			client.send <- pong
			continue
		}

		// NICK
		// https://modern.ircdocs.horse/#nick-message
		if parsed.Command == "NICK" {
			// has enough params?
			if len(parsed.Params) != 1 {
				//
				client.send <- fmt.Sprintf(":%s 461 * %s :Not enough parameters.", server.name, parsed.Command)
				continue
			}

			// validate nick
			ok := server.regex["nick"].MatchString(parsed.Params[0])
			if !ok {
				client.send <- fmt.Sprintf(":%s 432 * %s :Erroneus nickname.", server.name, parsed.Params[0])
				continue
			}

			_, exists := server.ClientByNickname(parsed.Params[0])
			if exists {
				client.send <- fmt.Sprintf(":%s 433 * %s :Nickname is already in use.", server.name, parsed.Params[0])
				continue
			}

			// set nickname
			client.SetNickname(parsed.Params[0])

			// check for handshake
			if !client.handshake {
				client.send <- fmt.Sprintf("NOTICE %s :AUTH :*** Your ID is: %s", client.nickname, client.id)
				client.send <- fmt.Sprintf("NOTICE %s :AUTH :*** Looking up your hostname..", client.nickname)

				// lookup address
				// todo: resolver
				addr, err := net.LookupAddr(client.IP())
				if err != nil {
					// if it cant be resolved use ip
					client.SetHostname(client.IP())
				} else {
					client.SetHostname(addr[0])
				}

				prefix := 4
				if strings.Count(client.IP(), ":") >= 2 {
					prefix = 6
				}

				// cloak
				client.SetHostname(fmt.Sprintf("ipv%d-%s.vhost", prefix, client.id))

				client.send <- fmt.Sprintf(":%s 001 %s :Welcome to the IRC network! 🎂", server.name, client.nickname)
				client.send <- fmt.Sprintf(":%s 002 %s :Your host is %s, running version -1", server.name, client.nickname, server.name)
				client.send <- fmt.Sprintf(":%s 376 %s :End of /MOTD command", server.name, client.nickname)
				client.handshake = true

				server.AddClient(client)
			}

			continue
		}

		// USER
		if parsed.Command == "USER" {
			if len(parsed.Params) < 4 {
				client.send <- fmt.Sprintf(":%s 461 %s %s :Not enough parameters.", server.name, client.nickname, strings.Join(parsed.Params, " "))
				continue
			}

			if client.username != "" {
				client.send <- fmt.Sprintf(":%s 462 %s :You may not reregister.", server.name, client.nickname)
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
			clientCount, channelCount := server.Stats()
			client.send <- fmt.Sprintf(":%s 461 %s :There are %d active users on %d channels.", server.name, client.nickname, clientCount, channelCount)

			var buf string
			clients := server.Clients()
			for _, c := range clients {
				buf = buf + fmt.Sprintf("%s ", c.nickname)
			}

			client.send <- fmt.Sprintf("NOTICE %s :DEBUG :*** %s", client.nickname, buf)
			continue
		}

		// JOIN
		if parsed.Command == "JOIN" {
			// join can have multiple channels separated by a comma
			targets := strings.Split(parsed.Params[0], ",")

			// todo: check and validate params - numeric

			for _, target := range targets {

				// ptr to existing channel or channel that will be created
				var channel *Channel

				// check if channel exists
				channel, exists := server.Channel(target)
				if !exists {
					// create if not
					channel = server.CreateChannel(target, client)
				}

				// add client to channel
				err = channel.AddClient(client, "")
				if err != nil {
					// todo: numeric
					log.Error().Err(err).Msgf("cant join channel %s", target)
					continue
				}

				// broadcast to all clients on the channel
				// that a client has joined
				channel.Broadcast(
					fmt.Sprintf(":%s JOIN %s", client.Prefix(), target),
					client,
					false,
				)

				topic := channel.Topic()
				if topic.text == "" {
					// send no topic
					client.send <- fmt.Sprintf(":%s 331 %s %s :no topic set",
						server.name,
						client.nickname,
						channel.name,
					)
				} else {
					// send topic if not empty
					client.send <- fmt.Sprintf(":%s 332 %s %s :%s",
						server.name,
						client.nickname,
						channel.name,
						topic.text,
					)
					// send time and author
					client.send <- fmt.Sprintf(":%s 329 %s %s %s %s %d",
						server.name,
						client.nickname,
						channel.name,
						topic.text,
						topic.author,
						topic.timestamp)
				}

				// get channel names (user list)
				names := channel.Names()

				// send names to client
				client.send <- fmt.Sprintf(":%s 353 %s %s %s :%s",
					server.name,
					client.nickname,
					"=",
					channel.name,
					names,
				)

			}

			continue
		}

		// PART
		if parsed.Command == "PART" {
			targets := strings.Split(parsed.Params[0], ",")

			for _, target := range targets {
				// try to get channel
				channel, exists := server.Channel(target)
				if !exists {
					client.send <- fmt.Sprintf(":%s 403 %s :No such channel.", server.name, target)
					continue
				}

				// remove client
				err := channel.RemoveClient(client)
				if err != nil {
					client.send <- fmt.Sprintf(":%s 462 %s :You're not on that channel.", server.name, target)
					continue
				}

				// broadcast that user has left the channel
				channel.Broadcast(fmt.Sprintf(":%s PART %s", client.Prefix(), target), client, false)
			}

			continue
		}

		// TOPIC
		if parsed.Command == "TOPIC" {
			target := parsed.Params[0]
			remainder := strings.Join(parsed.Params[1:len(parsed.Params)], " ")

			// todo: check and validate

			// try to get channel
			channel, exists := server.Channel(target)
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
				client, false)

			continue
		}

		// PRIVMSG
		// https://modern.ircdocs.horse/#privmsg-message
		if parsed.Command == "PRIVMSG" {
			targets := strings.Split(parsed.Params[0], ",")
			message := strings.Join(parsed.Params[1:len(parsed.Params)], " ")

			for _, target := range targets {

				// todo: check and validate target

				// is channel
				if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
					channel, exists := server.Channel(target)
					if !exists {
						client.send <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
							server.name,
							client.nickname)
						continue
					}
					// send message to channel
					channel.Broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s", client.Prefix(), channel.name, message), client, true)
					server.counters["ircd_channels_privmsg"].Inc()
					continue
				}
				// is user
				dest, exists := server.ClientByNickname(target)
				if !exists {
					client.send <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.name,
						client.nickname)
					continue
				}
				dest.send <- fmt.Sprintf(":%s PRIVMSG %s :%s", client.nickname, dest.nickname, message)
				continue
			}

		}

		// WHOIS
		if parsed.Command == "WHOIS" {
			target := parsed.Params[0]

			// todo: check and validate

			who, exists := server.ClientByNickname(target)
			if !exists {
				client.send <- fmt.Sprintf(
					":%s 401 %s :no such nick",
					server.name,
					client.nickname)
				continue
			}

			var buf string
			channels := server.Channels()
			for _, ch := range channels {
				if ch.IsMember(who) {
					prefix := "?"
					name := strings.Replace(ch.name, "#", "", 1)
					if ch.owner.nickname == who.nickname {
						prefix = "~"
					}
					buf = buf + fmt.Sprintf("%s%s ", prefix, name)
				}
			}

			client.send <- fmt.Sprintf(
				":%s 311 %s %s %s %s * :%s",
				server.name,
				client.nickname,
				who.nickname,
				who.username,
				who.hostname,
				who.realname,
			)

			client.send <- fmt.Sprintf(
				":%s 319 %s %s :%s",
				server.name,
				client.nickname,
				who.nickname,
				buf,
			)

			continue
		}
	}

	log.Info().Msg("client exited handle recv loop")
}
