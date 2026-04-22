package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyPasswordValid(t *testing.T) {
	pw := "correto-horse-battery-staple"
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(string(hash), pw) {
		t.Error("senha correta deveria validar")
	}
	if VerifyPassword(string(hash), "senha-errada") {
		t.Error("senha errada não deveria validar")
	}
}
