package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	_, prometheusEnabled := os.LookupEnv("PROMETHEUS")
	if prometheusEnabled {
		http.Handle("/metrics", promhttp.Handler())
	}
	go func() {
		log.Info().Msg("starting http, listening on :2112")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Error().Err(err).Msg("metrics http server stopped")
		}
	}()

	_, tlsEnabled := os.LookupEnv("TLS")

	config := ircd.ServerConfig{
		Name:     os.Getenv("SERVER_NAME"),
		Password: os.Getenv("SERVER_PASSWORD"),
		Network:  os.Getenv("NETWORK_NAME"),
		Version:  os.Getenv("SERVER_VERSION"),
		MOTD: []string{
			"4This is the message of the day.",
			"5It contains multiple lines because the lines could be long.",
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

			MaxTargets:     "PRIVMSG:3,WHOIS:1,JOIN:3",
			MaxTopicLength: 128,
			MaxUserLength:  20,
		},
	}

	server := ircd.NewServer(config)

	ctx, cancel := context.WithCancel(context.Background())

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatal().Err(err).Msg("cant listen")
	}
	go func() {
		log.Info().Msgf("starting irc, listening on tcp:%s", os.Getenv("PORT"))
		server.Serve(ctx, listener, false)
	}()

	var tlsListener net.Listener
	if config.TLS {
		tlsListener, err = tls.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT_TLS")),
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
		go func() {
			log.Info().Msgf("starting irc, listening on tcp:%s TLS", os.Getenv("PORT_TLS"))
			server.Serve(ctx, tlsListener, true)
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info().Msg("shutting down")
	cancel()
	listener.Close()
	if tlsListener != nil {
		tlsListener.Close()
	}
	server.Shutdown(10 * time.Second)
}
