package ircd

import "fmt"

func handleVersion(s *server, c clienter, m message) {
	if len(m.params) == 0 {
		c.sendRPL(s.name, rplVersion{
			client:   c.nickname(),
			version:  fmt.Sprintf("ircd-%s", s.version),
			server:   s.name,
			comments: "",
		})

		// todo: get tokens from config state and actually enforce them
		c.sendRPL(s.name, rplISupport{
			client: c.nickname(),
			tokens: "AWAYLEN=128 CASEMAPPING=ascii CHANLIMIT=#&:64 CHANNELLEN=50 CHANTYPES=#& ELIST=CUM HOSTLEN=128 KICKLEN=128 MODES=24 NICKLEN=31 PREFIX=(qaohv)~&@%+ TOPICLEN=307 USERLEN=18",
		})
		return
	}

	// todo: VERSION clients or other servers
	c.sendRPL(s.name, errNoSuchServer{
		client: c.nickname(),
		server: m.params[0],
	})
}
