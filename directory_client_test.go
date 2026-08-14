package ircd

import (
	"fmt"
	"sync"
	"testing"
)

func TestClientDirectoryRegisterAndLookup(t *testing.T) {
	d := newClientDirectory()
	th := newTestHandle("id1", "", "user", "host", false)
	d.register(th.sessionHandle)

	if _, ok := d.lookupID("id1"); !ok {
		t.Fatal("expected lookupID to find registered handle")
	}
	if _, ok := d.lookupNick("nick1"); ok {
		t.Fatal("expected lookupNick to fail before claimNick")
	}

	if err := d.claimNick("id1", "nick1"); err != nil {
		t.Fatalf("claimNick: %v", err)
	}
	h, ok := d.lookupNick("nick1")
	if !ok || h.id != "id1" {
		t.Fatalf("expected lookupNick to find id1, got %v %v", h, ok)
	}
}

func TestClientDirectoryClaimNickInUse(t *testing.T) {
	d := newClientDirectory()
	a := newTestHandle("a", "", "", "", false)
	b := newTestHandle("b", "", "", "", false)
	d.register(a.sessionHandle)
	d.register(b.sessionHandle)

	if err := d.claimNick("a", "shared"); err != nil {
		t.Fatalf("claimNick a: %v", err)
	}
	if err := d.claimNick("b", "shared"); err != errNickInUse {
		t.Fatalf("expected errNickInUse, got %v", err)
	}

	if err := d.claimNick("a", "shared"); err != nil {
		t.Fatalf("re-claiming own nick should succeed, got %v", err)
	}
}

func TestClientDirectoryClaimNickFreesOldNick(t *testing.T) {
	d := newClientDirectory()
	a := newTestHandle("a", "", "", "", false)
	d.register(a.sessionHandle)

	if err := d.claimNick("a", "first"); err != nil {
		t.Fatalf("claim first: %v", err)
	}
	if err := d.claimNick("a", "second"); err != nil {
		t.Fatalf("claim second: %v", err)
	}

	if _, ok := d.lookupNick("first"); ok {
		t.Fatal("expected old nick to be freed after rename")
	}
	if h, ok := d.lookupNick("second"); !ok || h.id != "a" {
		t.Fatal("expected new nick to resolve to a")
	}

	b := newTestHandle("b", "", "", "", false)
	d.register(b.sessionHandle)
	if err := d.claimNick("b", "first"); err != nil {
		t.Fatalf("expected freed nick to be claimable, got %v", err)
	}
}

func TestClientDirectoryUnregisterIsIdempotent(t *testing.T) {
	d := newClientDirectory()
	a := newTestHandle("a", "", "", "", false)
	d.register(a.sessionHandle)
	_ = d.claimNick("a", "nick")

	d.unregister("a")
	d.unregister("a")

	if _, ok := d.lookupID("a"); ok {
		t.Fatal("expected id to be gone after unregister")
	}
	if _, ok := d.lookupNick("nick"); ok {
		t.Fatal("expected nick to be freed after unregister")
	}
	if d.count() != 0 {
		t.Fatalf("expected count 0, got %d", d.count())
	}
}

func TestClientDirectoryConcurrentNickClaimIsExclusive(t *testing.T) {
	d := newClientDirectory()

	const n = 64
	handles := make([]*testHandle, n)
	for i := 0; i < n; i++ {
		handles[i] = newTestHandle(clientID(fmt.Sprintf("id%d", i)), "", "", "", false)
		d.register(handles[i].sessionHandle)
	}

	var wg sync.WaitGroup
	results := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = d.claimNick(handles[i].id, "contested")
		}(i)
	}
	wg.Wait()

	wins := 0
	for _, err := range results {
		if err == nil {
			wins++
		} else if err != errNickInUse {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if wins != 1 {
		t.Fatalf("expected exactly 1 winner, got %d", wins)
	}

	winner, ok := d.lookupNick("contested")
	if !ok {
		t.Fatal("expected contested nick to resolve")
	}
	found := false
	for i, err := range results {
		if err == nil && handles[i].id == winner.id {
			found = true
		}
	}
	if !found {
		t.Fatal("directory's winner does not match the goroutine that got a nil error")
	}
}
