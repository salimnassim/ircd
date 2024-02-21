package ircd

import (
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
		if parsed.Command == "WHOIS" {
			handleWhois(server, client, parsed)
			continue
		}

		// https://modern.ircdocs.horse/#who-message
		if parsed.Command == "WHO" {
			handleWho(server, client, parsed)
		}

	}

}
