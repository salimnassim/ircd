package ircd

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleNick(s *server, c *client, m message) {
	// nick params should be 1
	if len(m.params) != 1 {
		c.sendRPL(s.name, errNeedMoreParams{
			client:  c.nickname(),
			command: m.command,
		})
		return
	}

	// nicks should be more than 2 characters and less than 16
	if len(m.params[0]) < 2 || len(m.params[0]) > 16 {
		c.sendRPL(s.name, errErroneusNickname{
			client: c.nickname(),
			nick:   m.params[0],
		})
		return
	}

	// validate nickname
	ok := s.regex[regexKeyNick].MatchString(m.params[0])
	if !ok {
		c.sendRPL(s.name, errErroneusNickname{
			client: c.nickname(),
			nick:   m.params[0],
		})
		return
	}

	// check if nick is already in use
	_, ok = s.clients.get(m.params[0])
	if ok {
		c.sendRPL(s.name, errNicknameInUse{
			client: c.nickname(),
			nick:   m.params[0],
		})
		return
	}

	c.setNickname(m.params[0])

	// check for handshake
	if !c.handshake {

		// send handshake preamble
		c.sendNotice(noticeAuth{
			client:  c.nickname(),
			message: fmt.Sprintf("Your ID is: %s", c.id),
		})

		c.sendNotice(noticeAuth{
			client:  c.nickname(),
			message: fmt.Sprintf("Your IP address is: %s", c.ip),
		})

		c.sendNotice(noticeAuth{
			client:  c.nickname(),
			message: "Looking up your hostname...",
		})

		// lookup address
		// todo: resolver
		addr, err := net.LookupAddr(c.ip)
		if err != nil {
			// if it cant be resolved use ip
			c.setHostname(c.ip)
		} else {
			c.setHostname(addr[0])
		}

		// ipv4 or ipv6 connection for cloaking mask
		prefix := 4
		if strings.Count(c.ip, ":") > 1 {
			prefix = 6
		}

		c.sendRPL(s.name, rplWelcome{
			client:   c.nickname(),
			network:  "hello",
			hostname: c.prefix(),
		})

		c.sendRPL(s.name, rplYourHost{
			client:     c.nickname(),
			serverName: os.Getenv("SERVER_NAME"),
			version:    os.Getenv("SERVER_VERSION"),
		})

		// cloak
		c.setHostname(fmt.Sprintf("ipv%d-%s.vhost", prefix, c.id))
		c.sendNotice(noticeAuth{
			client:  c.nickname(),
			message: fmt.Sprintf("Your hostname has been cloaked to %s", c.host),
		})

		c.sendRPL(s.name, rplMotdStart{
			client: c.nickname(),
			server: s.name,
			text:   "MOTD -",
		})

		for _, line := range s.motd() {
			c.sendRPL(s.name, rplMotd{
				client: c.nickname(),
				text:   line,
			})
		}

		c.sendRPL(s.name, rplEndOfMotd{
			client: c.nickname(),
		})

		// set default modes
		c.addMode(modeClientInvisible)
		c.addMode(modeClientVhost)

		if s.tls {
			c.addMode(modeClientTLS)
		}

		c.send <- fmt.Sprintf("MODE %s %s", c.nickname(), c.modestring())

		c.handshake = true
	}
}
