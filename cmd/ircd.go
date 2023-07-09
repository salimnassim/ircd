package main

import (
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pyroscope-io/client/pyroscope"

	_ "net/http/pprof"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd"
)

func main() {

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	pyroscope.Start(pyroscope.Config{
		ApplicationName: "ircd",
		ServerAddress:   "http://pyroscope:4040",
		Logger:          nil,
		Tags:            map[string]string{"hostname": os.Getenv("HOSTNAME")},

		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	config := ircd.ServerConfig{
		Name: "ircd",
	}

	server := ircd.NewServer(config)

	listener, err := net.Listen("tcp", ":6667")
	if err != nil {
		log.Fatal().Err(err).Msg("unable to listen")
		os.Exit(1)
	}
	defer listener.Close()

	log.Info().Msg("starting http, listening on :2112")
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		http.ListenAndServe(":2112", nil)
		select {}
	}()

	log.Info().Msg("starting irc, listening on :6667")
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("unable to accept connection")
			continue
		}
		log.Info().Msgf("accepted connection from %s", connection.RemoteAddr())
		go ircd.HandleConnectionRead(connection, server)
	}

}
