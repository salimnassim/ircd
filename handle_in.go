package ircd

import (
	"github.com/rs/zerolog/log"
)

func handleConnectionIn(c *client, s *server) {
	defer func() {
		c.stop <- true
	}()

	for message := range c.recv {
		parsed, err := parseMessage(message)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse message in handler: %s", message)
			continue
		}

		log.Debug().Str("nick", c.nickname()).Msgf("%s", parsed.raw)

		// NICK
		// https://modern.ircdocs.horse/#nick-message
		if parsed.command == "NICK" {
			handleNick(s, c, parsed)
			continue
		}

		// USER
		// https://modern.ircdocs.horse/#user-message
		if parsed.command == "USER" {
			handleUser(s, c, parsed)
			continue
		}

		// LUSERS
		// https://modern.ircdocs.horse/#lusers-message
		if parsed.command == "LUSERS" {
			handleLusers(s, c, parsed)
			continue
		}

		// JOIN
		// https://modern.ircdocs.horse/#join-message
		if parsed.command == "JOIN" {
			handleJoin(s, c, parsed)
			continue
		}

		// PART
		// https://modern.ircdocs.horse/#part-message
		if parsed.command == "PART" {
			handlePart(s, c, parsed)
			continue
		}

		// TOPIC
		// https://modern.ircdocs.horse/#topic-message
		if parsed.command == "TOPIC" {
			handleTopic(s, c, parsed)
			continue
		}

		// PRIVMSG
		// https://modern.ircdocs.horse/#privmsg-message
		if parsed.command == "PRIVMSG" {
			handlePrivmsg(s, c, parsed)
			continue
		}

		// WHOIS
		// https://modern.ircdocs.horse/#whois-message
		if parsed.command == "WHOIS" {
			handleWhois(s, c, parsed)
			continue
		}

		// https://modern.ircdocs.horse/#who-message
		if parsed.command == "WHO" {
			handleWho(s, c, parsed)
			continue
		}

		// https://modern.ircdocs.horse/#mode-message
		if parsed.command == "MODE" {
			handleMode(s, c, parsed)
			continue
		}
	}
}
