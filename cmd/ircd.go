package main

import (
	"crypto/tls"
	"fmt"
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
		Name:    os.Getenv("SERVER_NAME"),
		Network: os.Getenv("NETWORK_NAME"),
		Version: os.Getenv("SERVER_VERSION"),
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

	go func(server ircd.Server, isTLS bool) {
		log.Info().Msgf("starting irc, listening on tcp:%s", os.Getenv("PORT"))
		listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT")))
		if err != nil {
			log.Fatal().Err(err).Msg("unable to listen")
		}
		server.Run(listener, isTLS)
		defer listener.Close()
	}(server, false)

	if config.TLS {
		go func(server ircd.Server, isTLS bool) {
			log.Info().Msgf("starting irc, listening on tcp:%s TLS", os.Getenv("PORT_TLS"))
			listener, err := tls.Listen(
				"tcp", fmt.Sprintf(":%s", os.Getenv("PORT_TLS")),
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
				log.Fatal().Err(err).Msg("unable to listen tls")
			}
			server.Run(listener, isTLS)
			defer listener.Close()
		}(server, true)
	}

	select {}
}
