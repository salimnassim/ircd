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
		client.sendRPL(server.name, errNeedMoreParams{
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
		client.sendRPL(server.name, errNicknameInUse{
			client: client.Nickname(),
			nick:   message.Params[0],
		})
		return
	}

	client.SetNickname(message.Params[0])

	// check for handshake
	if !client.handshake {

		// send handshake preamble
		client.sendNotice(noticeAuth{
			client:  client.Nickname(),
			message: fmt.Sprintf("Your ID is: %s", client.id),
		})

		client.sendNotice(noticeAuth{
			client:  client.Nickname(),
			message: fmt.Sprintf("Your IP address is: %s", client.ip),
		})

		client.sendNotice(noticeAuth{
			client:  client.Nickname(),
			message: "Looking up your hostname...",
		})

		// lookup address
		// todo: resolver
		addr, err := net.LookupAddr(client.ip)
		if err != nil {
			// if it cant be resolved use ip
			client.SetHostname(client.ip)
		} else {
			client.SetHostname(addr[0])
		}

		// ipv4 or ipv6 connection for cloaking mask
		prefix := 4
		if strings.Count(client.ip, ":") > 1 {
			prefix = 6
		}

		client.sendRPL(server.name, rplWelcome{
			client:   client.Nickname(),
			network:  "hello",
			hostname: client.Prefix(),
		})

		client.sendRPL(server.name, rplYourHost{
			client:     client.Nickname(),
			serverName: os.Getenv("SERVER_NAME"),
			version:    "0.1",
		})

		client.sendRPL(server.name, rplEndOfMotd{
			client: client.Nickname(),
		})

		// cloak
		client.SetHostname(fmt.Sprintf("ipv%d-%s.vhost", prefix, client.id))
		client.sendNotice(noticeAuth{
			client:  client.Nickname(),
			message: fmt.Sprintf("Your hostname has been cloaked to %s", client.hostname),
		})

		client.handshake = true
	}
}
