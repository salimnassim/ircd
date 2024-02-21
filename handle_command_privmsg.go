package ircd

import (
	"fmt"
	"strings"
)

func handlePrivmsg(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.sendRPL(server.name, errNotRegistered{
			client: client.Nickname(),
		})
		return
	}

	targets := strings.Split(message.Params[0], ",")
	text := strings.Join(message.Params[1:len(message.Params)], " ")

	for _, target := range targets {
		// is channel
		if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
			channel, exists := server.channels.Get(target)
			if !exists {
				client.sendRPL(server.name, errNoSuchChannel{
					client:  client.Nickname(),
					channel: target,
				})
				continue
			}

			// is user a member of the channel?
			if !server.channels.IsMember(client, channel) {
				client.sendRPL(server.name, errNotOnChannel{
					client:  client.Nickname(),
					channel: channel.name,
				})
				continue
			}

			// send message to channel
			channel.Broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s",
				client.Prefix(), channel.name, text), client.id, true)
			promPrivmsgChannel.Inc()
			continue
		}
		// is user
		dest, exists := server.clients.Get(target)
		if !exists {
			client.sendRPL(server.name, errNoSuchChannel{
				client:  client.Nickname(),
				channel: target,
			})
			continue
		}
		dest.send <- fmt.Sprintf(":%s PRIVMSG %s :%s",
			client.nickname, dest.nickname, text)
		promPrivmsgClient.Inc()
		continue
	}
}
