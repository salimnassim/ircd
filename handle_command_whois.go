package ircd

func handleWhois(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.sendRPL(server.name, errNotRegistered{
			client: client.Nickname(),
		})
		return
	}

	target := message.Params[0]
	who, exists := server.clients.Get(target)
	if who == nil || !exists {
		client.sendRPL(server.name, errNoSuchNick{
			client: client.Nickname(),
			nick:   target,
		})
		return
	}

	// https://modern.ircdocs.horse/#rplwhoisuser-311
	client.sendRPL(server.name, rplWhoisUser{
		client:   client.nickname,
		nick:     who.Nickname(),
		username: who.Username(),
		host:     who.Hostname(),
		realname: who.Realname(),
	})

	channels := []string{}
	memberOf := server.channels.MemberOf(who)
	for _, c := range memberOf {
		if !c.secret {
			channels = append(channels, c.name)
		}
	}

	client.sendRPL(server.name, rplWhoisChannels{
		client:   client.Nickname(),
		nick:     who.Nickname(),
		channels: channels,
	})
}
