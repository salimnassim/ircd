package ircd

import (
	"fmt"
	"strings"
)

func handleUser(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.send <- fmt.Sprintf(":%s 451 :You have not registered.",
			server.name)
		return
	}

	if len(message.Params) < 4 {
		client.send <- fmt.Sprintf(":%s 461 %s %s :Not enough parameters.",
			server.name, client.nickname, strings.Join(message.Params, " "))
		return
	}

	if client.username != "" {
		client.send <- fmt.Sprintf(":%s 462 %s :You may not reregister.",
			server.name, client.nickname)
		return
	}

	// todo: validate
	username := message.Params[0]
	realname := message.Params[3]

	client.SetUsername(username, realname)
}
