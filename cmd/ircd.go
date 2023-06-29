package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd"
)

func main() {

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		http.ListenAndServe(":2112", mux)
		select {}
	}()

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
	client.Hostname = strings.Split(client.Connection.LocalAddr().String(), ":")[0]

	go handleConnectionRead(client, server)
	go handleConnectionIn(client)
	go handleConnectionOut(client)
}

func handleConnectionOut(client *ircd.Client) {
	for message := range client.Out {
		n, err := client.Connection.Write([]byte(message + "\r\n"))
		if err != nil {
			log.Error().Err(err).Msgf("unable to write message to client (%s)", client.Connection.RemoteAddr())
			break
		}
		log.Info().Msgf("out(%5d)> %s", n, message)
	}
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
			nickname := split[1]

			_, err := server.GetClient(split[1])
			if err != nil {
				nicknext := fmt.Sprintf("%s-%d", nickname, server.GetRandomInt())
				client.Out <- fmt.Sprintf(":%s 433 * %s :Nickname is already in use. Your nickname has been changed to %s", server.Name, nickname, nicknext)
				nickname = nicknext
			}

			client.Nickname = nickname

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
			message := strings.Split(line, ":")[1]

			// to channel
			if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "?") || strings.HasPrefix(target, "~") {
				channel, err := server.GetChannel(target)
				if err != nil {
					client.Out <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.Name,
						client.Nickname)
					log.Error().Err(err).Msgf("privmsg channel %s does not exist", target)
					continue
				}
				channel.Broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s", client.GetTarget(), target, message), client, true)
				server.Counters["ircd_channels_privmsg"].Inc()
			} else {
				// to user
				tc, err := server.GetClient(target)
				if err != nil {
					client.Out <- fmt.Sprintf(":%s 401 %s :no such nick/channel",
						server.Name,
						client.Nickname)
					log.Error().Err(err).Msgf("privmsg channel %s does not exist", target)
					continue
				}
				tc.Out <- fmt.Sprintf(":%s PRIVMSG %s :%s", client.Nickname, tc.Nickname, message)
				server.Counters["ircd_clients_privmsg"].Inc()
			}
			continue
		}

		if strings.HasPrefix(line, "JOIN") {
			target := split[1]

			channel, err := server.GetChannel(target)
			if err != nil && channel == nil {
				channel = server.CreateChannel(target)
			}
			err = channel.AddClient(client, "")
			if err != nil {
				log.Error().Err(err).Msgf("cant join channel %s", target)
				continue
			}

			channel.Broadcast(
				fmt.Sprintf(":%s JOIN %s",
					client.GetTarget(), target),
				client,
				false,
			)

			topic := channel.GetTopic()
			if topic.Topic == "" {
				client.Out <- fmt.Sprintf(":%s 331 %s %s :no topic set",
					server.Name,
					client.Nickname,
					channel.Name,
				)
			} else {
				client.Out <- fmt.Sprintf(":%s 332 %s %s :%s",
					server.Name,
					client.Nickname,
					channel.Name,
					topic.Topic,
				)
			}

			continue
		}

		if strings.HasPrefix(line, "PART") {
			target := split[1]

			channel, err := server.GetChannel(target)
			if err != nil {
				log.Error().Err(err).Msgf("tried to part channel that does not exist %s", target)
			}

			channel.Broadcast(fmt.Sprintf(":%s PART %s", client.GetTarget(), target), client, false)
			channel.RemoveClient(client)
		}

		if strings.HasPrefix(strings.ToUpper(line), "LUSERS") {
			counts := server.Counts()
			client.Out <- fmt.Sprintf(":%s 251 %s :There are %d users and %d invisible on %d servers",
				server.Name,
				client.Nickname,
				counts.Active,
				counts.Invisible,
				1,
			)
			client.Out <- fmt.Sprintf(":%s 254 %s %d :channels formed",
				server.Name,
				client.Nickname,
				counts.Channels)
			continue
		}

		if strings.HasPrefix(line, "TOPIC") {
			target := split[1]
			message := strings.Replace(split[2], ":", "", 1)

			channel, err := server.GetChannel(target)
			if err != nil {
				log.Error().Err(err).Msgf("tried to change topic %s", target)
			}

			channel.SetTopic(message, client.Nickname)
			channel.Broadcast(
				fmt.Sprintf(":%s 332 %s %s :%s",
					server.Name,
					client.Nickname,
					channel.Name,
					message),
				client, false)
			continue
		}

		if strings.HasPrefix(line, "QUIT") {
			break
		}

	}

	log.Info().Msgf("closing client from connection read (%s)", client.Connection.RemoteAddr())
	client.Close()
	server.RemoveClient(client)
}
