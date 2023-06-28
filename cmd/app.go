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

	server := ircd.NewServer("servername")

	log.Info().Msg("starting server, listening on :6667")

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

	server.AddClient(client)

	go handleConnectionRead(client, server)
	go handleConnectionIn(client)
	go handleConnectionOut(client)
}

func handleConnectionOut(client *ircd.Client) {
	for message := range client.Out {
		n, err := client.Connection.Write([]byte(message + "\r\n"))
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.Connection.RemoteAddr())
			continue
		}
		log.Info().Msgf("out(%5d)> %s", n, message)
	}

	client.Connection.Close()
}

func handleConnectionIn(client *ircd.Client) {
	for message := range client.In {
		log.Info().Msgf(" in(%5d)> %s", len(message), message)
	}
}

func handleConnectionRead(client *ircd.Client, server *ircd.Server) {
	reader := bufio.NewReader(client.Connection)
	for {
		line, err := reader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		if err != nil {
			log.Error().Err(err).Msgf("unable to read from client (%s)", client.Connection.RemoteAddr())
			break
		}

		client.In <- line
		split := strings.Split(line, " ")

		if strings.HasPrefix(line, "PING") {
			pong := strings.Replace(line, "PING", "PONG", 1)
			client.Out <- pong
			continue
		}

		if strings.HasPrefix(line, "NICK") {
			if _, nok := server.GetClient(split[1]); nok {
				client.Out <- fmt.Sprintf(":%s %s 433 :Nickname is already in use", server.Name, client.Nickname)
				continue
			}

			client.Nickname = split[1]

			if !client.Handshake {
				client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Looking up your hostname..", client.Nickname)
				client.Out <- fmt.Sprintf("NOTICE %s :AUTH :*** Checking ident", client.Nickname)
				client.Out <- fmt.Sprintf(":%s 001 %s :Welcome to the IRC network!", server.Name, client.Nickname)
				client.Out <- fmt.Sprintf(":%s 376 %s :End of /MOTD command", server.Name, client.Nickname)
				client.Handshake = true
			}
			continue
		}

		if strings.HasPrefix(line, "USER") {
			client.Username = split[1]
			continue
		}

		if strings.HasPrefix(line, "PRIVMSG") {
			target := split[1]
			message := strings.Replace(split[2], ":", "", 1)

			if strings.HasPrefix(target, "?#~") {
				channel, err := server.GetChannel(target)
				if err != nil {
					log.Error().Err(err).Msgf("channel %s does not exist", target)
					continue
				}
				log.Debug().Msgf("channel is %s", channel.Name)
				channel.Broadcast(message)
			}

			// todo: user to user
			continue
		}

		if strings.HasPrefix(line, "JOIN") {
			target := split[1]

			channel, err := server.GetChannel(target)
			if err != nil && channel == nil {
				channel = server.CreateChannel(target)

			}
			err = channel.Join(client, "")
			if err != nil {
				log.Error().Err(err).Msgf("cant join channel %s", target)
				continue
			}

			channel.Broadcast(
				fmt.Sprintf(":%s!%s@%s JOIN %s",
					client.Nickname, client.Username, client.Connection.RemoteAddr(), target),
			)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(line), "LUSERS") {
			log.Debug().Msgf("%v", server.Clients)
			log.Debug().Msgf("%v", server.Channels)

			client.Out <- fmt.Sprintf("%s :There are %d users and %d invisible on %d servers",
				client.Nickname,
				len(server.Clients),
				len(server.Clients),
				1,
			)
			continue
		}

		if strings.HasPrefix(line, "QUIT") {
			log.Debug().Msgf("%v", server.Channels)
			log.Debug().Msgf("%v", server.Clients)
			break
		}

	}

	log.Info().Msgf("closing client from connection read (%s)", client.Connection.RemoteAddr())

	client.Close()
	server.RemoveClient(client)
}
