package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Clients = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ircd",
		Name:      "clients",
		Help:      "Number of connected clients.",
	})

	Channels = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ircd",
		Name:      "channels",
		Help:      "Number of existing channels.",
	})

	Command = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "command",
		Help:      "Number of commands received.",
	}, []string{"name"})
)
