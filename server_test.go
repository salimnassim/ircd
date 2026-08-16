package ircd

import (
	"context"
	"net"
	"testing"
	"time"
)

func testServerConfig() ServerConfig {
	return ServerConfig{
		Name:           "test.server",
		Network:        "TestNet",
		Version:        "0.0.0-test",
		MOTD:           []string{"line one"},
		CloakSecret:    []byte("test-cloak-secret"),
		PingFrequency:  30,
		PongMaxLatency: 10,
	}
}

func startTestServerObj(t *testing.T) (*Server, string) {
	t.Helper()
	return startTestServerWithConfig(t, testServerConfig())
}

func startTestServerWithConfig(t *testing.T, config ServerConfig) (*Server, string) {
	t.Helper()

	srv := NewServer(config)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go srv.Serve(ctx, listener, false)

	t.Cleanup(func() {
		cancel()
		listener.Close()
	})

	return srv, listener.Addr().String()
}

func TestServerShutdownNotifiesAndDrainsConnectedSessions(t *testing.T) {
	srv, addr := startTestServerObj(t)

	alice := dialTestClient(t, addr)
	defer alice.conn.Close()
	alice.register("alice")

	if got := srv.clients.count(); got != 1 {
		t.Fatalf("expected 1 registered session before shutdown, got %d", got)
	}

	start := time.Now()
	done := make(chan struct{})
	go func() {
		srv.Shutdown(2 * time.Second)
		close(done)
	}()

	alice.expect("ERROR :Closing Link: Server shutting down")
	alice.expectClosed()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected Shutdown to return once the connected session drained")
	}
	if elapsed := time.Since(start); elapsed >= 2*time.Second {
		t.Fatalf("expected shutdown to drain well before its timeout, took %v", elapsed)
	}

	if got := srv.clients.count(); got != 0 {
		t.Fatalf("expected no sessions left registered after shutdown, got %d", got)
	}
}

func TestServerShutdownReturnsAfterTimeoutWhenASessionWontDrain(t *testing.T) {
	srv := NewServer(testServerConfig())

	th := newTestHandle("stuck", "stuck", "user", "host", false)
	var terminateCalled bool
	th.terminateFn = func(error) { terminateCalled = true }
	srv.clients.register(th.sessionHandle)

	start := time.Now()
	srv.Shutdown(100 * time.Millisecond)
	elapsed := time.Since(start)

	if !terminateCalled {
		t.Fatal("expected Shutdown to call terminate on the stuck session")
	}
	if elapsed < 100*time.Millisecond {
		t.Fatalf("expected Shutdown to wait out its full drain timeout, returned after %v", elapsed)
	}
	if elapsed > 1*time.Second {
		t.Fatalf("expected Shutdown to be bounded by its timeout instead of hanging, took %v", elapsed)
	}
}
