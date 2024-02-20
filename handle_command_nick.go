package ircd

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleNick(server *Server, client *Client, message Message) {
	if len(message.Params) != 1 {
		client.send <- fmt.Sprintf(":%s 461 * %s :Not enough parameters.",
			server.name, message.Command)
		return
	}

	// nicks should be more than 2 characters and less than 9
	if len(message.Params[0]) < 2 || len(message.Params[0]) > 9 {
		client.send <- fmt.Sprintf(":%s 432 * %s :Erroneus nickname.",
			server.name, message.Params[0])
		return
	}

	// validate nickname
	ok := server.regex["nick"].MatchString(message.Params[0])
	if !ok {
		client.send <- fmt.Sprintf(":%s 432 * %s :Erroneus nickname.",
			server.name, message.Params[0])
		return
	}

	// check if nick is already in use
	_, exists := server.clients.GetByNickname(message.Params[0])
	if exists {
		client.send <- fmt.Sprintf(":%s 433 * %s :Nickname is already in use.",
			server.name, message.Params[0])
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
