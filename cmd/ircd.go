package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	config := ircd.ServerConfig{
		Name: "ircd",
		MOTD: []string{
			"This is the message of the day.",
			"It contains multiple lines because the lines could be long.",
			"🍩🍫🍡🍦🍬🍮",
		},
	}

	server := ircd.NewServer(config)

	log.Info().Msg("starting irc, listening on :6667")

	listener, err := net.Listen("tcp", ":6667")
	if err != nil {
		log.Fatal().Err(err).Msg("unable to listen")
		os.Exit(1)
	}
	defer listener.Close()

	go func() {
		log.Info().Msg("starting http, listening on :2112")

		_, ok := os.LookupEnv("PROMETHEUS")
		if ok {
			http.Handle("/metrics", promhttp.Handler())
		}
		http.ListenAndServe(":2112", nil)
	}()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("unable to accept connection")
			continue
		}
		log.Info().Msgf("accepted connection from %s", connection.RemoteAddr())
		go ircd.HandleConnection(connection, server)
	}
}
