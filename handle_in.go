package ircd

import (
	"bufio"
	"strings"
)

func handleConnectionIn(c *client, s *server) {
	reader := bufio.NewReader(c.conn)
	scanner := bufio.NewScanner(reader)

	alive := true
	for alive {
		select {
		case <-c.killIn:
			alive = false
		default:
			scanner.Scan()
			if scanner.Err() != nil {
				if c.quitReason() == "" {
					c.kill("EOF")
				}
				continue
			}
			line := strings.Trim(
				scanner.Text(), "\r\n",
			)
			parsed, err := parseMessage(line)
			if err != nil {
				continue
			}
			s.router.handle(s, c, parsed)
		}
	}

	c.conn.Close()
}
