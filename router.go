package ircd

import (
	"sync"
)

var nilHandler handlerFunc = func(s *server, c *client, m message) {}

type handlerFunc func(s *server, c *client, m message)
type middlewareFunc func(s *server, c *client, m message, next handlerFunc) handlerFunc

type router interface {
	// Register cmd route, assign optional middleware.
	registerHandler(cmd string, h handlerFunc, mws ...middlewareFunc)
	// Execute handler.
	handle(s *server, c *client, m message) error
}

type commandRouter struct {
	mu *sync.RWMutex

	server *server

	handlers   map[string]handlerFunc
	middleware map[string][]middlewareFunc
}

func NewCommandRouter(s *server) *commandRouter {
	return &commandRouter{
		mu: &sync.RWMutex{},

		server:     s,
		handlers:   map[string]handlerFunc{},
		middleware: map[string][]middlewareFunc{},
	}
}

func (cr *commandRouter) registerHandler(cmd string, h handlerFunc, mws ...middlewareFunc) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.handlers[cmd] = h
	cr.middleware[cmd] = mws
}

func (cr *commandRouter) handle(s *server, c *client, m message) error {
	cr.mu.RLock()
	h, ok := cr.handlers[m.command]
	if !ok {
		return errorCommandNotFound
	}
	mws := cr.middleware[m.command]
	cr.mu.RUnlock()

	cr.wrap(s, c, m, h, mws)(cr.server, c, m)
	return nil
}

func (cr *commandRouter) wrap(s *server, c *client, m message, handler handlerFunc, middleware []middlewareFunc) handlerFunc {
	if handler == nil {
		return nil
	}
	wrap := handler
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		wrap = mw(s, c, m, wrap)
		if wrap == nil {
			return nilHandler
		}
	}
	return wrap
}
