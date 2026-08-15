package ircd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	clientGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ircd",
		Name:      "clients",
		Help:      "Number of connected clients.",
	})

	channelGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ircd",
		Name:      "channels",
		Help:      "Number of existing channels.",
	})

	commandCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "command",
		Help:      "Number of commands received.",
	}, []string{"name"})

	connectionsRejectedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "connections_rejected",
		Help:      "Number of connections rejected before session start.",
	}, []string{"reason"})
)
