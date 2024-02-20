package ircd

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleNick(server *Server, client *Client, message Message) {
	// nick params should be 1
	if len(message.Params) != 1 {
		client.sendRPL(server.name, rplNeedMoreParams{
			client:  client.Nickname(),
			command: message.Command,
		})
		return
	}

	// nicks should be more than 2 characters and less than 16
	if len(message.Params[0]) < 2 || len(message.Params[0]) > 16 {
		client.sendRPL(server.name, rplErroneusNickname{
			client: client.Nickname(),
			nick:   message.Params[0],
		})
		return
	}

	// validate nickname
	ok := server.regex["nick"].MatchString(message.Params[0])
	if !ok {
		client.sendRPL(server.name, rplErroneusNickname{
			client: client.Nickname(),
			nick:   message.Params[0],
		})
		return
	}

	// check if nick is already in use
	_, exists := server.clients.GetByNickname(message.Params[0])
	if exists {
		client.sendRPL(server.name, rplNicknameInUse{
			client: client.Nickname(),
			nick:   message.Params[0],
		})
		return
	}

	client.SetNickname(message.Params[0])

	// check for handshake
	if !client.handshake {
		client.send <- fmt.Sprintf("NOTICE %s :AUTH :*** Your ID is: %s",
			client.nickname, client.id)
		client.send <- fmt.Sprintf("NOTICE %s :AUTH :*** Your IP address is: %s",
			client.nickname, client.ip)
		client.send <- fmt.Sprintf("NOTICE %s :AUTH :*** Looking up your hostname...",
			client.nickname)

		// lookup address
		// todo: resolver
		addr, err := net.LookupAddr(client.ip)
		if err != nil {
			// if it cant be resolved use ip
			client.SetHostname(client.ip)
		} else {
			client.SetHostname(addr[0])
		}

		prefix := 4
		if strings.Count(client.ip, ":") > 1 {
			prefix = 6
		}

		// cloak
		client.SetHostname(fmt.Sprintf("ipv%d-%s.vhost", prefix, client.id))
		client.send <- fmt.Sprintf("NOTICE %s :AUTH :*** Your hostname has been cloaked to %s",
			client.nickname, client.hostname)
		client.send <- fmt.Sprintf(":%s 001 %s :Welcome to the IRC network! 🎂",
			server.name, client.nickname)
		client.send <- fmt.Sprintf(":%s 002 %s :Your host is %s (%s), running version -1",
			server.name, client.nickname, server.name, os.Getenv("HOSTNAME"))
		client.send <- fmt.Sprintf(":%s 376 %s :End of /MOTD command",
			server.name, client.nickname)
		client.handshake = true
	}
}
