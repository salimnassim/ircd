package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Number of connected clients.
	Clients = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ircd",
		Name:      "clients",
		Help:      "Number of connected clients",
	})
	Command = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "command",
		Help:      "Numbers of commands received.",
	}, []string{"name"})
	// Number of existing channels.
	Channels = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ircd",
		Name:      "channels",
		Help:      "Number of existing channels",
	})
)
