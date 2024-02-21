package ircd

import (
	"fmt"
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handlePart(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.sendRPL(server.name, errNotRegistered{
			client: client.Nickname(),
		})
		return
	}

	targets := strings.Split(message.Params[0], ",")

	for _, target := range targets {
		// try to get channel
		channel, exists := server.channels.Get(target)
		if !exists {
			client.sendRPL(server.name, errNoSuchChannel{
				client:  client.Nickname(),
				channel: target,
			})
			continue
		}

		// remove client
		channel.RemoveClient(client)

		// broadcast that user has left the channel
		channel.Broadcast(fmt.Sprintf(":%s PART %s :<no reason given>",
			client.Prefix(), target), client.id, false)

		if channel.clients.Count() == 0 {
			server.channels.Delete(channel.name)
			metrics.Channels.Dec()
		}
	}
}
