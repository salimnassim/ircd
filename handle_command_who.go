package ircd

import (
	"strings"
)

func handleWho(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.sendRPL(server.name, errNotRegistered{
			client: client.Nickname(),
		})
		return
	}

	if len(message.Params) == 0 {
		client.sendRPL(server.name, errNeedMoreParams{
			client:  client.Nickname(),
			command: message.Command,
		})
		return
	}

	target := message.Params[0]
	if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
		channel, ok := server.channels.Get(target)
		if !ok {
			client.sendRPL(server.name, errNoSuchChannel{
				client:  client.Nickname(),
				channel: target,
			})
			return
		}

		for _, c := range channel.clients.All() {
			client.sendRPL(server.name, rplWhoReply{
				client:   client.Nickname(),
				channel:  channel.name,
				username: c.Username(),
				host:     c.Hostname(),
				server:   server.name,
				nick:     c.Nickname(),
				flags:    "",
				hopcount: 0,
				realname: c.Realname(),
			})
		}

		client.sendRPL(server.name, rplEndOfWho{
			client: client.Nickname(),
			mask:   "",
		})
		return
	}

	// todo: support querying users with mask
}
