package ircd

import (
	"testing"
	"time"
)

func TestConnLimiterPerIPRejectsExtraConnections(t *testing.T) {
	cfg := testServerConfig()
	cfg.MaxConnectionsPerIP = 2
	_, addr := startTestServerWithConfig(t, cfg)

	alice := dialTestClient(t, addr)
	defer alice.conn.Close()
	alice.register("alice")

	bob := dialTestClient(t, addr)
	defer bob.conn.Close()
	bob.register("bob")

	carol := dialTestClient(t, addr)
	defer carol.conn.Close()
	carol.expect("ERROR :Closing Link: too many connections from your host")
	carol.expectClosed()
}

func TestConnLimiterGlobalRejectsExtraConnections(t *testing.T) {
	cfg := testServerConfig()
	cfg.MaxConnectionsGlobal = 1
	_, addr := startTestServerWithConfig(t, cfg)

	alice := dialTestClient(t, addr)
	defer alice.conn.Close()
	alice.register("alice")

	bob := dialTestClient(t, addr)
	defer bob.conn.Close()
	bob.expect("ERROR :Closing Link: server is at its connection limit")
	bob.expectClosed()
}

func TestConnLimiterReleasesSlotOnDisconnect(t *testing.T) {
	cfg := testServerConfig()
	cfg.MaxConnectionsPerIP = 1
	srv, addr := startTestServerWithConfig(t, cfg)

	alice := dialTestClient(t, addr)
	alice.register("alice")
	alice.conn.Close()

	deadline := time.Now().Add(2 * time.Second)
	for srv.clients.count() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := srv.clients.count(); got != 0 {
		t.Fatalf("expected alice's session to clean up before the deadline, clients.count() = %d", got)
	}

	bob := dialTestClient(t, addr)
	defer bob.conn.Close()
	bob.register("bob")
}
