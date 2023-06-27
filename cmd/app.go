package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd"
)

func main() {
	listener, err := net.Listen("tcp", ":6667")
	if err != nil {
		log.Fatal().Err(err).Msg("unable to listen")
		os.Exit(1)
	}
	defer listener.Close()

	server := ircd.NewServer()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("unable to accept connection")
			continue
		}
		log.Info().Msgf("accepted connection from %s", connection.RemoteAddr())
		go handleConnection(server, connection)
	}

}

func handleConnection(server *ircd.Server, connection net.Conn) {
	log.Info().Msgf("handling connection")

	client, err := ircd.NewClient(connection)
	if err != nil {
		log.Error().Err(err).Msg("unable to create client")
		return
	}

	go handleConnectionInbox(client)
	go handleConnectionOutbox(client)
	go handleConnectionRead(client)
}

func handleConnectionOutbox(client *ircd.Client) {
	writer := bufio.NewWriter(client.Connection)
	for message := range client.Out {
		n, err := writer.WriteString(message)
		if err != nil {
			log.Error().Err(err).Msg("unable to write message to client")
			break
		}
		log.Info().Msgf("out(%5d)> %s", n, message)
	}
}

func handleConnectionInbox(client *ircd.Client) {
	for message := range client.In {
		log.Info().Msgf(" in(%5d)> %s", len(message), message)
	}
}

func handleConnectionRead(client *ircd.Client) {
	defer client.Connection.Close()
	reader := bufio.NewReader(client.Connection)
	for {
		line, err := reader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		if err != nil {
			log.Error().Err(err).Msg("cant read from connection")
			break
		}

		client.In <- line

		if strings.HasPrefix(line, "PING") {
			pong := strings.Replace(line, "PING", "PONG", 1)
			client.Out <- pong
		}

		split := strings.Split(line, " ")
		if strings.HasPrefix(line, "NICK") {
			client.Nickname = split[1]

			if !client.Handshake {
				client.Out <- ircd.Line("NOTICE", "AUTH :*** Looking up your hostname...")
				client.Out <- ircd.Line("NOTICE", "AUTH :*** Checking ident")
				client.Out <- ircd.Line("001", fmt.Sprintf("ircd - welcome to the server %s", client.Nickname))
				client.Out <- ircd.Line("376", "End of /MOTD command")
				client.Handshake = true
			}
		}

		if strings.HasPrefix(line, "USER") {
			client.Username = split[1]
		}

	}
}
