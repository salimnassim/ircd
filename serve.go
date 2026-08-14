package ircd

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func acceptLoop(ctx context.Context, listener net.Listener, isTLS bool, deps sessionDeps) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			log.Error().Err(err).Msg("unable to accept connection")
			continue
		}
		id := clientID(uuid.Must(uuid.NewRandom()).String())
		go runSession(ctx, conn, id, isTLS, deps)
	}
}
