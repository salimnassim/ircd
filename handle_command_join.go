package ircd

import (
	"fmt"
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handleJoin(s *server, c *client, m Message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	// join can have multiple channels separated by a comma
	targets := strings.Split(m.Params[0], ",")
	for _, target := range targets {
		// channels have to start with # or & and be less than 9 charaacters
		if !strings.HasPrefix(target, "#") && !strings.HasPrefix(target, "&") || len(target) > 9 {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// validate channel name
		ok := s.regex["channel"].MatchString(target)
		if !ok {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// ptr to existing channel or channel that will be created
		var channel *channel

		channel, exists := s.channels.Get(target)
		if !exists {
			// create channel if it does not exist
			channel = NewChannel(target, c.id)

			// todo: use channel.id instead of target
			s.channels.Add(target, channel)

			metrics.Channels.Inc()
		}

		// add client to channel
		err := channel.addClient(c, "")
		if err != nil {
			c.sendRPL(s.name, errBadChannelKey{
				client:  c.nickname(),
				channel: channel.name,
			})
			continue
		}

		// broadcast to all clients on the channel
		// that a client has joined
		channel.broadcast(
			fmt.Sprintf(":%s JOIN %s", c.prefix(), channel.name),
			c.id,
			false,
		)

		// chanowner
		if channel.owner == c.id {
			channel.broadcast(
				fmt.Sprintf(":%s MODE %s +o %s", s.name, channel.name, c.nick),
				c.id,
				false,
			)
		}

		topic := channel.topic()
		if topic.text == "" {
			// send no topic
			c.sendRPL(s.name, rplNoTopic{
				client:  c.nickname(),
				channel: channel.name,
			})
		} else {
			// send topic if not empty
			c.sendRPL(s.name, rplTopic{
				client:  c.nickname(),
				channel: channel.name,
				topic:   topic.text,
			})

			// send time and author
			c.sendRPL(s.name, rplTopicWhoTime{
				client:  c.nickname(),
				channel: channel.name,
				nick:    topic.author,
				setat:   topic.timestamp,
			})
		}

		// get channel names (user list)
		names := channel.names()

		// send names to client
		symbol := "="
		c.sendRPL(s.name, rplNamReply{
			client:  c.nickname(),
			symbol:  symbol,
			channel: channel.name,
			nicks:   names,
		})

		c.sendRPL(s.name, rplEndOfNames{
			client:  c.nickname(),
			channel: channel.name,
		})
	}

}
