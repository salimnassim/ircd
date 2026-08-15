package ircd

import (
	"context"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"
)

func acceptLoop(ctx context.Context, listener net.Listener, isTLS bool, deps sessionDeps) {
	var backoff time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if backoff == 0 {
				backoff = 5 * time.Millisecond
			} else if backoff *= 2; backoff > time.Second {
				backoff = time.Second
			}
			slog.Error("unable to accept connection", "err", err, "backoff", backoff)
			time.Sleep(backoff)
			continue
		}
		backoff = 0

		ip := remoteHost(conn)
		if ok, reason := deps.limiter.acquire(ip); !ok {
			connectionsRejectedCounter.WithLabelValues(reason).Inc()
			rejectConnection(conn, reason)
			continue
		}

		id := clientID(uuid.Must(uuid.NewRandom()).String())
		go runSession(ctx, conn, id, isTLS, deps)
	}
}

func rejectConnection(conn net.Conn, reason string) {
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, _ = io.WriteString(conn, errorCommand{text: rejectionMessage(reason)}.command()+"\r\n")
	conn.Close()
}

func rejectionMessage(reason string) string {
	switch reason {
	case "global":
		return "server is at its connection limit"
	case "rate":
		return "connecting too quickly, try again shortly"
	default: // "per_ip"
		return "too many connections from your host"
	}
}
