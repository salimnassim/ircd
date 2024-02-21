package ircd

import "github.com/salimnassim/ircd/metrics"

func handleLusers(server *Server, client *Client, message Message) {
	clients, channels := server.Stats()

	client.sendRPL(server.name, rplLuserClient{
		client:    client.Nickname(),
		users:     clients,
		invisible: 0,
		servers:   1,
	})
	client.sendRPL(server.name, rplLuserOp{
		client: client.Nickname(),
		ops:    0,
	})
	client.sendRPL(server.name, rplLuserChannels{
		client:   client.Nickname(),
		channels: channels,
	})

	metrics.Lusers.Inc()
}
