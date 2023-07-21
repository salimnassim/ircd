package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pyroscope-io/client/pyroscope"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd"
)

func main() {

	if os.Getenv("PYROSCOPE_ENABLE") != "" {
		log.Info().Msg("pyroscope is enabled")

		runtime.SetMutexProfileFraction(5)
		runtime.SetBlockProfileRate(5)

		pyroscope.Start(pyroscope.Config{
			ApplicationName: "ircd",
			ServerAddress:   os.Getenv("PYROSCOPE_ADDRESS"),
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
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	config := ircd.ServerConfig{
		Name: "ircd",
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
		if os.Getenv("PROMETHEUS_ENABLE") != "" {
			log.Info().Msg("prometheus is enabled")
			http.Handle("/metrics", promhttp.Handler())
		}
		http.HandleFunc("/", server.IndexHandler)

		http.ListenAndServe(":2112", nil)
	}()

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
