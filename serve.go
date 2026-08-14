package ircd

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
			log.Error().Err(err).Dur("backoff", backoff).Msg("unable to accept connection")
			time.Sleep(backoff)
			continue
		}
		backoff = 0

		id := clientID(uuid.Must(uuid.NewRandom()).String())
		go runSession(ctx, conn, id, isTLS, deps)
	}
}
