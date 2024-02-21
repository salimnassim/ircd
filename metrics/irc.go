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
	// Number of existing channels.
	Channels = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ircd",
		Name:      "channels",
		Help:      "Number of existing channels",
	})
	// Number of PING messages.
	Pings = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "pings",
		Help:      "Number of PING messages",
	})
	// Number of PRIVMSG sent to channels.
	PrivmsgChannel = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "privmsg_channel",
		Help:      "Number of PRIVMSG sent to channels",
	})
	// Number of PRIVMSG sent to clients.
	PrivmsgClient = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "privmsg_client",
		Help:      "Number of PRIVMSG sent to clients",
	})
)
