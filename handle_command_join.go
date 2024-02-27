package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handleJoin(s *server, c clienter, m message) {
	// join can have multiple channels separated by a comma
	targets := strings.Split(m.params[0], ",")

	keys := []string{}
	// keys set?
	if len(m.params) >= 2 {
		keys = strings.Split(m.params[1], ",")
		// number of target has to match with number of keys
		if len(targets) != len(keys) {
			c.sendRPL(s.name, errNeedMoreParams{
				client:  c.nickname(),
				command: m.command,
			})
			return
		}
	}

	for i, target := range targets {
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
			ch.addMode(modeChannelRestrictTopic)

			metrics.Channels.Inc()
		}

		// if channel has +z, do not allow joining without tls
		if ch.hasMode(modeChannelTLSOnly) && !c.tls() {
			c.sendCommand(noticeCommand{
				client:  c.nickname(),
				message: "Cannot join channel (+z)",
			})
			return
		}

		// is channel invite only, and client does not have an invitation?
		if ch.hasMode(modeChannelInviteOnly) && !ch.isInvited(c) {
			c.sendRPL(s.name, errInviteOnlyChan{
				client:  c.nickname(),
				channel: ch.name(),
			})
			return
		} else {
			// remove from invite map if invite is accepted
			ch.removeInvite(c.id())
		}

		// if channel has key, compare key
		if ch.hasMode(modeChannelKey) {
			if len(keys) < i+1 {
				c.sendRPL(s.name, errBadChannelKey{
					client:  c.nickname(),
					channel: ch.name(),
				})
				continue
			}
			if ch.key() != keys[i] {
				c.sendRPL(s.name, errBadChannelKey{
					client:  c.nickname(),
					channel: ch.name(),
				})
				continue
			}
		}

		// add client to channel
		ch.clients().add(c)

		// broadcast to all clients on the channel
		// that a client has joined
		ch.broadcastCommand(joinCommand{
			prefix:  c.prefix(),
			channel: ch.name(),
		}, c.id(), false)

		// chanowner
		if ch.owner() == c.id() {
			ch.clients().addMode(c, modeMemberOwner)
			ch.broadcastCommand(modeCommand{
				source:     s.name,
				target:     ch.name(),
				modestring: ch.clients().modestring(c),
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

		// send current channel modes to client
		c.sendCommand(modeCommand{
			source:     s.name,
			target:     ch.name(),
			modestring: ch.modestring(),
			args:       "",
		})
	}
}
