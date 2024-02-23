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

		switch parsed.command {
		case "PING":
			handlePing(s, c, parsed)
			continue
		case "NICK":
			handleNick(s, c, parsed)
			continue
		case "USER":
			handleUser(s, c, parsed)
			continue
		case "LUSERS":
			handleLusers(s, c, parsed)
			continue
		case "JOIN":
			handleJoin(s, c, parsed)
			continue
		case "PART":
			handlePart(s, c, parsed)
			continue
		case "TOPIC":
			handleTopic(s, c, parsed)
			continue
		case "PRIVMSG":
			handlePrivmsg(s, c, parsed)
			continue
		case "WHOIS":
			handleWhois(s, c, parsed)
			continue
		case "WHO":
			handleWho(s, c, parsed)
			continue
		case "MODE":
			handleMode(s, c, parsed)
			continue
		case "QUIT":
			handleQuit(s, c, parsed)
			continue
		case "DEBUG":
			// breakpoint here
			continue
		}
	}
}
