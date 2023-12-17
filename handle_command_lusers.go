package ircd

import "fmt"

func handleLusers(server *Server, client *Client, message Message) string {
	clientCount, channelCount := server.Stats()
	return fmt.Sprintf(":%s 251 %s :There are %d active users on %d channels.",
		server.name, client.nickname, clientCount, channelCount)
}
