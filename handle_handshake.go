package ircd

import (
	"fmt"
	"net"
	"strings"
)

func handleHandshake(s *server, c *client) {
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

		c.sendRPL(s.name, rplWelcome{
			client:   c.nickname(),
			network:  s.network,
			hostname: c.prefix(),
		})

		c.sendRPL(s.name, rplYourHost{
			client:     c.nickname(),
			serverName: s.name,
			version:    s.version,
		})

		// ipv4 or ipv6 connection for cloaking mask
		prefix := 4
		if strings.Count(c.ip, ":") > 1 {
			prefix = 6
		}

		tls := "plain"
		if c.tls {
			tls = "tls"
		}

		// cloak
		c.setHostname(fmt.Sprintf("ipv%d-%s-%s.vhost", prefix, tls, c.id))
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

		if c.tls {
			c.addMode(modeClientTLS)
		}

		c.send <- fmt.Sprintf("MODE %s %s", c.nickname(), c.modestring())

		c.handshake = true
	}
}
