package ircd

import (
	"io"
)

func handleConnectionOut(c *client) {
	alive := true
	for alive {
		select {
		case <-c.killOut:
			alive = false
		case m := <-c.out:
			_, err := io.WriteString(c.conn, m+"\r\n")
			if err != nil {
				c.kill("Broken pipe")
				continue
			}
		}
	}

	c.conn.Close()
}
