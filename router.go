package ircd

import "sync"

type handlerFunc func(s *server, c *client, m message)
type middlewareFunc func(next handlerFunc) handlerFunc

type router interface {
	add(cmd string, h handlerFunc)
	use(m ...middlewareFunc)
	match(cmd string) error
	subrouter(r *router)
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
	cr.handlers[cmd] = h
	cr.middleware[cmd] = mws
	cr.mu.Unlock()
}

func (cr *commandRouter) handle(c *client, m message) error {
	cr.mu.RLock()
	h, ok := cr.handlers[m.command]
	if !ok {
		return errorCommandNotFound
	}
	mws := cr.middleware[m.command]
	cr.mu.RUnlock()

	cr.wrap(h, mws)(cr.server, c, m)
	return nil
}

func (cr *commandRouter) wrap(handler handlerFunc, middleware []middlewareFunc) handlerFunc {
	if handler == nil {
		return nil
	}
	wrap := handler
	for _, m := range middleware {
		wrap = m(wrap)
	}
	return wrap
}
