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

		c.sendRPL(s.name, rplISupport{
			client: c.nickname(),
			tokens: "AWAYLEN=128 CASEMAPPING=ascii CHANLIMIT=#&:64 CHANNELLEN=50 CHANTYPES=#& HOSTLEN=128 KICKLEN=128 MODES=24 NICKLEN=31 PREFIX=(qaohv)~&@%+ TOPICLEN=307 USERLEN=18",
		})
		return
	}

	c.sendRPL(s.name, errNoSuchServer{
		client: c.nickname(),
		server: m.params[0],
	})
}
