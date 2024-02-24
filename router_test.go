package ircd

import (
	"net"
	"testing"
	"time"
)

type connTester struct {
	deadline time.Time
}

func (c *connTester) Read(b []byte) (n int, err error) {
	return 0, nil
}
func (c *connTester) SetDeadline(t time.Time) error {
	c.deadline = t
	return nil
}
func (c *connTester) SetReadDeadline(t time.Time) error {
	c.deadline = t
	return nil
}
func (c *connTester) SetWriteDeadline(t time.Time) error {
	c.deadline = t
	return nil
}
func (c *connTester) Write(b []byte) (int, error) { return 0, nil }
func (c *connTester) Close() error                { return nil }
func (C *connTester) LocalAddr() net.Addr         { return &net.IPAddr{} }
func (C *connTester) RemoteAddr() net.Addr        { return &net.IPAddr{} }

func Test(t *testing.T) {

	s := NewServer(ServerConfig{
		Name: "mock",
	})

	c, _ := newClient(&connTester{}, "test")

	t.Run("access handler function", func(t *testing.T) {
		m := message{
			command: "TEST",
		}
		want := "hello"
		got := ""
		r := NewCommandRouter(s)
		r.registerHandler("TEST", func(s *server, c *client, m message) {
			got = "hello"
		})
		err := r.handle(s, c, m)
		if err != nil {
			t.Error(err)
		}
		if want != got {
			t.Errorf("got: %s, want %s", got, want)
		}
	})

	t.Run("access handler function with middleware", func(t *testing.T) {
		m := message{
			command: "TEST",
		}
		want := "beforeafter"
		got := ""
		router := NewCommandRouter(s)

		router.registerHandler("TEST", func(s *server, c *client, m message) {
			got = got + "after"
		}, func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			got = "before"
			return next
		})

		err := router.handle(s, c, m)
		if err != nil {
			t.Error(err)
		}

		if want != got {
			t.Errorf("got: %s, want %s", got, want)
		}
	})

	t.Run("access handler function with nil middleware", func(t *testing.T) {
		m := message{
			command: "TEST",
		}
		want := "hello"
		got := ""
		router := NewCommandRouter(s)

		router.registerHandler("TEST", func(s *server, c *client, m message) {
			got = "hello"
		}, nil)

		err := router.handle(s, c, m)
		if err != nil {
			t.Error(err)
		}

		if want != got {
			t.Errorf("got: %s, want %s", got, want)
		}
	})

	t.Run("access handler function with multiple middleware", func(t *testing.T) {
		m := message{
			command: "TEST",
		}
		want := "onetwothree"
		got := ""
		router := NewCommandRouter(s)

		router.registerHandler("TEST", func(s *server, c *client, m message) {
			got = got + "three"
		}, func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			got = got + "one"
			return next
		}, func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			got = got + "two"
			return next
		})

		err := router.handle(s, c, m)
		if err != nil {
			t.Error(err)
		}

		if want != got {
			t.Errorf("got: %s, want %s", got, want)
		}
	})

	t.Run("access message from middleware", func(t *testing.T) {
		m := message{
			command: "TEST",
			params:  []string{"params0"},
		}
		want := m.params[0]
		got := ""
		router := NewCommandRouter(s)

		router.registerHandler("TEST", func(s *server, c *client, m message) {
			// nothing
		}, func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			got = m.params[0]
			return next
		})

		err := router.handle(s, c, m)
		if err != nil {
			t.Error(err)
		}

		if want != got {
			t.Errorf("got: %s, want %s", got, want)
		}
	})

	t.Run("exit early from middleware", func(t *testing.T) {
		m := message{
			command: "TEST",
		}
		want := "exit here"
		got := ""
		router := NewCommandRouter(s)

		router.registerHandler("TEST", func(s *server, c *client, m message) {
			// nothing
		}, func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			got = "exit here"
			return nil
		}, func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			got = "should not be here"
			return next
		})

		err := router.handle(s, c, m)
		if err != nil {
			t.Error(err)
		}

		if want != got {
			t.Errorf("got: %s, want %s", got, want)
		}
	})

	t.Run("access global middleware", func(t *testing.T) {
		m := message{
			command: "TEST",
		}
		want := "global"
		got := ""
		router := NewCommandRouter(s)

		router.registerGlobalMiddleware(func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			got = "global"
			return next
		})

		router.registerHandler("TEST", func(s *server, c *client, m message) {
			// nothing
		}, func(s *server, c *client, m message, next handlerFunc) handlerFunc {
			return next
		})

		err := router.handle(s, c, m)
		if err != nil {
			t.Error(err)
		}

		if want != got {
			t.Errorf("got: %s, want %s", got, want)
		}
	})
}
