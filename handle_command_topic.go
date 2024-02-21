package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handleTopic(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.sendRPL(server.name, errNotRegistered{
			client: client.Nickname(),
		})
		return
	}

	target := message.Params[0]

	// channel name max length is 50, check for allowed channel prefixes
	if !(strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&")) {
		client.sendRPL(server.name, errNoSuchChannel{
			client:  client.Nickname(),
			channel: target,
		})
		return
	}

	// try to get channel
	channel, exists := server.channels.Get(target)
	if !exists {
		client.sendRPL(server.name, errNoSuchChannel{
			client:  client.Nickname(),
			channel: target,
		})
		return
	}

	// set topic
	remainder := strings.Join(message.Params[1:len(message.Params)], " ")
	channel.SetTopic(remainder, client.nickname)
	metrics.Topic.Inc()

	// get topic
	topic := channel.Topic()

	// broadcast new topic to clients on channel
	channel.BroadcastRPL(
		rplTopic{
			client:  client.Nickname(),
			channel: channel.name,
			topic:   topic.text,
		},
		client.id,
		false,
	)
}
