package ircd

import "fmt"

func handleLusers(server *Server, client *Client, message Message) {
	clients, channels := server.Stats()

	client.send <- fmt.Sprintf(":%s 251 %s :There are %d users and %d invisible on %d servers.",
		server.name, client.nickname, clients, 0, 1)
	client.send <- fmt.Sprintf(":%s 252 %s %d :operator(s) online.",
		server.name, client.nickname, 0)
	client.send <- fmt.Sprintf(":%s 254 %s %d :channel(s) formed.",
		server.name, client.nickname, channels)
}
