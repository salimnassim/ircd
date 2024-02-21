package ircd

func handleUser(server *Server, client *Client, message Message) {
	if !client.handshake {
		client.sendRPL(server.name, errNotRegistered{
			client: client.Nickname(),
		})
		return
	}

	if len(message.Params) < 4 {
		client.sendRPL(server.name, errNeedMoreParams{
			client: client.nickname,
		})
		return
	}

	if client.username != "" {
		client.sendRPL(server.name, errAlreadyRegistered{
			client: client.Nickname(),
		})
		return
	}

	username := message.Params[0]
	realname := message.Params[3]

	client.SetUsername(username, realname)
}
