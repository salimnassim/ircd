package ircd

import "testing"

func TestCloakHostDeterministic(t *testing.T) {
	secret := []byte("shared-secret")

	a := cloakHost(secret, "203.0.113.7")
	b := cloakHost(secret, "203.0.113.7")

	if a != b {
		t.Fatalf("expected same address to produce the same cloak, got %q and %q", a, b)
	}
}

func TestCloakHostDiffersByAddress(t *testing.T) {
	secret := []byte("shared-secret")

	a := cloakHost(secret, "203.0.113.7")
	b := cloakHost(secret, "203.0.113.8")

	if a == b {
		t.Fatalf("expected different addresses to produce different cloaks, both were %q", a)
	}
}

func TestCloakHostDiffersBySecret(t *testing.T) {
	a := cloakHost([]byte("secret-one"), "203.0.113.7")
	b := cloakHost([]byte("secret-two"), "203.0.113.7")

	if a == b {
		t.Fatalf("expected different secrets to produce different cloaks, both were %q", a)
	}
}
