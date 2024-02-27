package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

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
		Name:     os.Getenv("SERVER_NAME"),
		Password: os.Getenv("SERVER_PASSWORD"),
		Network:  os.Getenv("NETWORK_NAME"),
		Version:  os.Getenv("SERVER_VERSION"),
		MOTD: []string{
			"\u00034This is the message of the day.\u0003",
			"\u00035It contains multiple lines because the lines could be long.\u0003",
			"🍩🍫🍡🍦🍬🍮",
		},
		TLS:             tlsEnabled,
		CertificateFile: os.Getenv("TLS_CERTIFICATE"),
		CertificateKey:  os.Getenv("TLS_KEY"),
		PingFrequency:   30,
		PongMaxLatency:  10,
		Parameters: ircd.ServerConfigParameters{
			MaxAwayLength:     128,
			CaseMapping:       "ascii",
			ChannelLimit:      "#&:64",
			ChannelModes:      "b,f,lk,ztSsrOmMiCc",
			MaxChannelLength:  50,
			ChannelTypes:      "&#",
			EList:             "",
			Excepts:           "",
			MaxHostnameLength: 32,
			MaxKickLength:     32,
			MaxList:           "b:16",
			MaxModes:          16,
			Network:           "Network",
			MaxNickLength:     31,
			ChannelPrefixes:   "(qaohv)~&@%+",
			// StatusMessage:     "~&@%+",
			MaxTargets:     "PRIVMSG:3,WHOIS:1,JOIN:3",
			MaxTopicLength: 128,
			MaxUserLength:  20,
		},
	}

	server := ircd.NewServer(config)

	go func(server ircd.Serverer, isTLS bool) {
		log.Info().Msgf("starting irc, listening on tcp:%s", os.Getenv("PORT"))
		listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT")))
		if err != nil {
			log.Fatal().Err(err).Msg("cant listen")
		}
		server.Run(listener, isTLS)
		defer listener.Close()
	}(server, false)

	if config.TLS {
		go func(server ircd.Serverer, isTLS bool) {
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
				log.Fatal().Err(err).Msg("cant listen tls")
			}
			server.Run(listener, isTLS)
			defer listener.Close()
		}(server, true)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
