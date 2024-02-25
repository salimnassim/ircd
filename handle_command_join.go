package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handleJoin(s *server, c clienter, m message) {
	// join can have multiple channels separated by a comma
	targets := strings.Split(m.params[0], ",")
	for _, target := range targets {
		// channels have to start with # or & and be less than 9 charaacters
		if !m.isTargetChannel() {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// validate channel name
		ok := s.regex[regexChannel].MatchString(target)
		if !ok {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// ptr to existing ch or ch that will be created
		var ch channeler

		ch, exists := s.Channels.get(target)
		if !exists {
			// create channel if it does not exist
			ch = newChannel(target, c.id())

			// todo: use channel.id instead of target
			s.Channels.add(ch.name(), ch)

			// set default channel modes
			ch.addMode(modeChannelNoExternal)

			metrics.Channels.Inc()
		}

		// if channel has +z, do not allow joining without tls
		if ch.hasMode(modeChannelTLSOnly) && !c.tls() {
			c.sendRPL(s.name, errNeedTLSJoin{
				client:  c.nickname(),
				channel: ch.name(),
			})
			return
		}

		// add client to channel
		err := ch.addClient(c, "")
		if err != nil {
			c.sendRPL(s.name, errBadChannelKey{
				client:  c.nickname(),
				channel: ch.name(),
			})
			continue
		}

		// broadcast to all clients on the channel
		// that a client has joined
		ch.broadcastCommand(joinCommand{
			prefix:  c.prefix(),
			channel: ch.name(),
		}, c.id(), false)

		// chanowner
		if ch.owner() == c.id() {
			ch.broadcastCommand(modeCommand{
				source:     s.name,
				target:     ch.name(),
				modestring: "+o",
				args:       c.nickname(),
			}, c.id(), false)
		}

		topic := ch.topic()
		if topic.text == "" {
			// send no topic
			c.sendRPL(s.name, rplNoTopic{
				client:  c.nickname(),
				channel: ch.name(),
			})
		} else {
			// send topic if not empty
			c.sendRPL(s.name, rplTopic{
				client:  c.nickname(),
				channel: ch.name(),
				topic:   topic.text,
			})

			// send time and author
			c.sendRPL(s.name, rplTopicWhoTime{
				client:  c.nickname(),
				channel: ch.name(),
				nick:    topic.author,
				setat:   topic.timestamp,
			})
		}

		// get channel names (user list)
		names := ch.names()

		// send names to client
		symbol := "="
		c.sendRPL(s.name, rplNamReply{
			client:  c.nickname(),
			symbol:  symbol,
			channel: ch.name(),
			nicks:   names,
		})

		c.sendRPL(s.name, rplEndOfNames{
			client:  c.nickname(),
			channel: ch.name(),
		})
	}

}
