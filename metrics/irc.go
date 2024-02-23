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
	Ping = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "ping",
		Help:      "Number of PING messages",
	})
	// Number of PONG replies.
	Pong = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "pong",
		Help:      "Number of PONG messages",
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
	Topic = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "topic",
		Help:      "Number of TOPIC messages",
	})
	Lusers = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "ircd",
		Name:      "lusers",
		Help:      "Number of LUSERS messages",
	})
)
