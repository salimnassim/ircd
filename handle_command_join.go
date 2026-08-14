package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handleJoin(s *server, c clienter, m message) {
	targets := strings.Split(m.params[0], ",")

	keys := []string{}

	if len(m.params) >= 2 {
		keys = strings.Split(m.params[1], ",")

		if len(targets) != len(keys) {
			c.sendRPL(s.name, errNeedMoreParams{
				client:  c.nickname(),
				command: m.command,
			})
			return
		}
	}

	for i, target := range targets {
		if !m.isTargetChannel() {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		ok := s.regex[regexChannel].MatchString(target)
		if !ok {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		var ch channeler

		ch, exists := s.Channels.get(target)
		if !exists {
			ch = newChannel(target, c.id())

			s.Channels.add(ch.name(), ch)

			ch.addMode(modeChannelNoExternal)
			ch.addMode(modeChannelRestrictTopic)

			metrics.Channels.Inc()
		}

		if ch.hasMode(modeChannelTLSOnly) && !c.tls() {
			c.sendCommand(noticeCommand{
				client:  c.nickname(),
				message: "Cannot join channel (+z)",
			})
			return
		}

		if ch.hasMode(modeChannelInviteOnly) && !ch.isInvited(c) {
			c.sendRPL(s.name, errInviteOnlyChan{
				client:  c.nickname(),
				channel: ch.name(),
			})
			return
		} else {
			ch.removeInvite(c.id())
		}

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

		ch.clients().add(c)

		ch.broadcastCommand(joinCommand{
			prefix:  c.prefix(),
			channel: ch.name(),
		}, c.id(), false)

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
			c.sendRPL(s.name, rplNoTopic{
				client:  c.nickname(),
				channel: ch.name(),
			})
		} else {
			c.sendRPL(s.name, rplTopic{
				client:  c.nickname(),
				channel: ch.name(),
				topic:   topic.text,
			})

			c.sendRPL(s.name, rplTopicWhoTime{
				client:  c.nickname(),
				channel: ch.name(),
				nick:    topic.author,
				setat:   topic.timestamp,
			})
		}

		names := ch.names()

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

		c.sendCommand(modeCommand{
			source:     s.name,
			target:     ch.name(),
			modestring: ch.modestring(),
			args:       "",
		})
	}
}
