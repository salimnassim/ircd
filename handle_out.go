package ircd

import (
	"io"

	"github.com/rs/zerolog/log"
)

func handleConnectionOut(c *client, s *server) {
	alive := true
	for alive {
		select {
		case <-c.killOut:
			log.Debug().Str("nick", c.nickname()).Msg("Killed out")
			alive = false
		case m := <-c.out:
			_, err := io.WriteString(c.conn, m+"\r\n")
			if err != nil {
				c.kill("Broken pipe")
				return
			}
		}
	}

	c.conn.Close()
}
