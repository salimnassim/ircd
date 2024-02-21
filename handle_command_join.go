package ircd

import (
	"fmt"
	"strings"
)

func handleJoin(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.sendRPL(server.name, errNotRegistered{
			client: client.Nickname(),
		})
		return
	}

	// join can have multiple channels separated by a comma
	targets := strings.Split(message.Params[0], ",")
	for _, target := range targets {
		// channels have to start with # or & and be less than 9 charaacters
		if !strings.HasPrefix(target, "#") && !strings.HasPrefix(target, "&") || len(target) > 9 {
			client.sendRPL(server.name, errNoSuchChannel{
				client:  client.Nickname(),
				channel: target,
			})
			continue
		}

		// validate channel name
		ok := server.regex["channel"].MatchString(target)
		if !ok {
			client.sendRPL(server.name, errNoSuchChannel{
				client:  client.Nickname(),
				channel: target,
			})
			continue
		}

		// ptr to existing channel or channel that will be created
		var channel *Channel

		channel, exists := server.channels.GetByName(target)
		if !exists {
			// create channel if it does not exist
			channel = NewChannel(target, client.id)

			// todo: use channel.id instead of target
			server.channels.Add(target, channel)

			promChannels.Inc()
		}

		// add client to channel
		err := channel.AddClient(client, "")
		if err != nil {
			client.sendRPL(server.name, errBadChannelKey{
				client:  client.Nickname(),
				channel: channel.name,
			})
			continue
		}

		// broadcast to all clients on the channel
		// that a client has joined
		channel.Broadcast(
			fmt.Sprintf(":%s JOIN %s", client.Prefix(), channel.name),
			client.id,
			false,
		)

		// chanowner
		if channel.owner == client.id {
			channel.Broadcast(
				fmt.Sprintf(":%s MODE %s +o %s", server.name, channel.name, client.nickname),
				client.id,
				false,
			)
		}

		topic := channel.Topic()
		if topic.text == "" {
			// send no topic
			client.sendRPL(server.name, rplNoTopic{
				client:  client.Nickname(),
				channel: channel.name,
			})
		} else {
			// send topic if not empty
			client.sendRPL(server.name, rplTopic{
				client:  client.Nickname(),
				channel: channel.name,
				topic:   topic.text,
			})

			// send time and author
			client.sendRPL(server.name, rplTopicWhoTime{
				client:  client.Nickname(),
				channel: channel.name,
				nick:    topic.author,
				setat:   topic.timestamp,
			})
		}

		// get channel names (user list)
		names := channel.Names()

		// send names to client
		symbol := "="
		client.sendRPL(server.name, rplNamReply{
			client:  client.Nickname(),
			symbol:  symbol,
			channel: channel.name,
			nicks:   names,
		})

		client.sendRPL(server.name, rplEndOfNames{
			client:  client.Nickname(),
			channel: channel.name,
		})
	}

}
