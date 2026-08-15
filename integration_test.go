package ircd

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

type testIRCClient struct {
	t    *testing.T
	conn net.Conn
	r    *bufio.Reader
}

func dialTestClient(t *testing.T, addr string) *testIRCClient {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return &testIRCClient{t: t, conn: conn, r: bufio.NewReader(conn)}
}

func (c *testIRCClient) send(line string) {
	c.t.Helper()
	c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := c.conn.Write([]byte(line + "\r\n")); err != nil {
		c.t.Fatalf("write %q: %v", line, err)
	}
}

func (c *testIRCClient) readLine() (string, error) {
	c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	line, err := c.r.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

func (c *testIRCClient) expect(substr string) string {
	c.t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := c.readLine()
		if err != nil {
			c.t.Fatalf("expecting %q, read error: %v", substr, err)
		}
		if strings.Contains(line, substr) {
			return line
		}
	}
	c.t.Fatalf("timed out waiting for line containing %q", substr)
	return ""
}

func (c *testIRCClient) expectClosed() {
	c.t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, err := c.readLine()
		if err != nil {
			return
		}
	}
	c.t.Fatal("expected connection to be closed by the server")
}

func (c *testIRCClient) register(nick string) {
	c.t.Helper()
	c.send("NICK " + nick)
	c.send("USER " + nick + " 0 0 :" + nick + " Realname")
	c.expect("376") // RPL_ENDOFMOTD
}

func startTestServer(t *testing.T) (addr string, deps sessionDeps, cancel context.CancelFunc) {
	t.Helper()

	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps = testSessionDeps(cd, chd)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go acceptLoop(ctx, listener, false, deps)

	t.Cleanup(func() {
		cancel()
		listener.Close()
	})

	return listener.Addr().String(), deps, cancel
}
func TestIntegrationJoinPrivmsgQuitKill(t *testing.T) {
	addr, deps, _ := startTestServer(t)
	deps.operators.add("root", "secret")

	alice := dialTestClient(t, addr)
	defer alice.conn.Close()
	alice.register("alice")

	bob := dialTestClient(t, addr)
	defer bob.conn.Close()
	bob.register("bob")

	// JOIN: alice joins first, sees her own JOIN and empty NAMES.
	alice.send("JOIN #test")
	alice.expect("JOIN #test")
	alice.expect("366 alice #test") // RPL_ENDOFNAMES

	// bob joins the same channel; both alice and bob should observe it.
	bob.send("JOIN #test")
	bob.expect("JOIN #test")
	alice.expect("JOIN #test") // broadcast of bob's join to existing members

	// PRIVMSG to the channel: bob should see alice's message, alice should not.
	alice.send("PRIVMSG #test :hello channel")
	bob.expect("PRIVMSG #test :hello channel")

	// PRIVMSG directly to a user.
	alice.send("PRIVMSG bob :direct hello")
	bob.expect("PRIVMSG bob :direct hello")

	// OPER + KILL: a third connection authenticates as an operator and kills bob.
	root := dialTestClient(t, addr)
	defer root.conn.Close()
	root.register("root")
	root.send("OPER root secret")
	root.expect("381") // RPL_YOUREOPER

	root.send("KILL bob :spamming")
	bob.expect("Killed by root (spamming)")
	bob.expectClosed()

	// alice should see bob's forced disconnect reflected as a QUIT in the channel.
	alice.expect("QUIT :Killed by root (spamming)")

	// QUIT: alice disconnects herself cleanly.
	alice.send("QUIT :goodbye")
	alice.expect("Quit: goodbye")
	alice.expectClosed()
}

func TestIntegrationChannelBanSetListAndEnforced(t *testing.T) {
	_, addr := startTestServerObj(t)

	alice := dialTestClient(t, addr)
	defer alice.conn.Close()
	alice.register("alice")

	alice.send("JOIN #test")
	alice.expect("JOIN #test")

	alice.send("MODE #test +b mallory!*@*")
	alice.expect("MODE #test +b mallory!*@*")

	alice.send("MODE #test b")
	alice.expect("367 alice #test mallory!*@*")
	alice.expect("368")

	mallory := dialTestClient(t, addr)
	defer mallory.conn.Close()
	mallory.send("NICK mallory")
	mallory.send("USER mallory 0 0 :Mallory")
	mallory.expect("376")

	mallory.send("JOIN #test")
	mallory.expect("474")

	alice.send("MODE #test -b mallory!*@*")
	alice.expect("MODE #test -b mallory!*@*")

	mallory.send("JOIN #test")
	mallory.expect("JOIN #test")
}
