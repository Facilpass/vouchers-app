package auth

import (
	"testing"
	"time"
)

func TestSessionRoundtrip(t *testing.T) {
	secret := "super-secret-key-64-chars-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	m := NewSessionManager(secret, 8*time.Hour)

	token, err := m.Sign("admin")
	if err != nil {
		t.Fatal(err)
	}
	user, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if user != "admin" {
		t.Errorf("user = %q, want admin", user)
	}
}

func TestSessionTampered(t *testing.T) {
	m := NewSessionManager("secret-key-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", time.Hour)
	token, _ := m.Sign("admin")

	tampered := token[:len(token)-1] + "X"
	_, err := m.Verify(tampered)
	if err == nil {
		t.Error("token adulterado deveria falhar verify")
	}
}

func TestSessionExpired(t *testing.T) {
	m := NewSessionManager("secret-key-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", -time.Hour)
	token, _ := m.Sign("admin")
	_, err := m.Verify(token)
	if err == nil {
		t.Error("token expirado deveria falhar")
	}
}

func TestSessionDifferentSecret(t *testing.T) {
	m1 := NewSessionManager("secret-a-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", time.Hour)
	m2 := NewSessionManager("secret-b-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", time.Hour)
	token, _ := m1.Sign("admin")
	_, err := m2.Verify(token)
	if err == nil {
		t.Error("token com secret diferente deveria falhar")
	}
}
