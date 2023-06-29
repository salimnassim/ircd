package main

import (
	"net"
	"net/http"
	"os"

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

	server := ircd.NewServer(os.Getenv("SERVER_NAME"))

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

	go ircd.HandleConnectionRead(client, server)
	go ircd.HandleConnectionIn(client)
	go ircd.HandleConnectionOut(client)
}
