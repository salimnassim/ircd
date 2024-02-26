package ircd

import (
	"fmt"
	"net"
	"strings"
)

func handleHandshake(s *server, c clienter) {
	if !c.handshake() {
		// send handshake preamble
		c.sendCommand(noticeCommand{
			client:  c.nickname(),
			message: fmt.Sprintf("AUTH :*** Your ID is: %s", c.id()),
		})

		c.sendCommand(noticeCommand{
			client:  c.nickname(),
			message: fmt.Sprintf("AUTH :*** Your IP address is: %s", c.ip()),
		})

		c.sendCommand(noticeCommand{
			client:  c.nickname(),
			message: "AUTH :*** Looking up your hostname...",
		})

		// lookup address
		// todo: resolver
		addr, err := net.LookupAddr(c.ip())
		if err != nil {
			// if it cant be resolved use ip
			c.setHostname(c.ip())
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
		if strings.Count(c.ip(), ":") > 1 {
			prefix = 6
		}

		tls := "plain"
		if c.tls() {
			tls = "tls"
		}

		// cloak
		c.setHostname(fmt.Sprintf("ipv%d-%s-%s.vhost", prefix, tls, c.id()))
		c.sendCommand(noticeCommand{
			client:  c.nickname(),
			message: fmt.Sprintf("AUTH :*** Your hostname has been cloaked to %s", c.hostname()),
		})

		c.sendCommand(noticeCommand{
			client: c.nickname(),
			message: fmt.Sprintf(
				"NOTICE :*** Server is listening on ports %s",
				strings.Join(s.ports, ", "),
			),
		})

		c.sendRPL(s.name, rplMotdStart{
			client: c.nickname(),
			server: s.name,
			text:   "MOTD -",
		})

		for _, line := range s.MOTD() {
			c.sendRPL(s.name, rplMotd{
				client: c.nickname(),
				text:   line,
			})
		}

		c.sendRPL(s.name, rplEndOfMotd{
			client: c.nickname(),
		})

		c.sendRPL(s.name, rplISupport{
			client: c.nickname(),
			tokens: s.params,
		})

		// set default modes
		c.addMode(modeClientVhost)

		if c.tls() {
			c.addMode(modeClientTLS)
		}

		c.sendCommand(modeCommand{
			target:     c.nickname(),
			modestring: c.modestring(),
			args:       "",
		})

		c.setHandshake(true)
	}
}
