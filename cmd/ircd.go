package main

import (
	"crypto/tls"
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

	go func() {
		log.Info().Msg("starting http, listening on :2112")

		_, ok := os.LookupEnv("PROMETHEUS")
		if ok {
			http.Handle("/metrics", promhttp.Handler())
		}
		http.ListenAndServe(":2112", nil)
	}()

	_, tlsEnabled := os.LookupEnv("TLS")

	config := ircd.ServerConfig{
		Name: "ircd",
		MOTD: []string{
			"This is the message of the day.",
			"It contains multiple lines because the lines could be long.",
			"🍩🍫🍡🍦🍬🍮",
		},
		TLS:             tlsEnabled,
		CertificateFile: os.Getenv("TLS_CERTIFICATE"),
		CertificateKey:  os.Getenv("TLS_KEY"),
	}

	server := ircd.NewServer(config)

	// go func(listener net.Listener, server ircd.Server) {
	// 	server.Run(listener)
	// }(listener, server)

	var listener net.Listener
	var err error
	if !config.TLS {
		log.Info().Msg("starting irc, listening on tcp:6667")
		listener, err = net.Listen("tcp", ":6667")
		if err != nil {
			log.Fatal().Err(err).Msg("unable to listen")
		}
	}

	if config.TLS {
		log.Info().Msg("starting irc, listening on tcp:6697 TLS")
		listener, err = tls.Listen(
			"tcp", ":6697",
			&tls.Config{
				GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
					cert, err := tls.LoadX509KeyPair(config.CertificateFile, config.CertificateKey)
					if err != nil {
						return nil, err
					}
					return &cert, nil
				},
			})
		if err != nil {
			log.Fatal().Err(err).Msg("unable to listen")
		}
	}

	defer listener.Close()

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
