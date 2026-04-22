package ui

import "testing"

func TestCSRFRoundtrip(t *testing.T) {
	c := NewCSRFManager("secret-key-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	sid := "session-id-abc"
	token := c.Generate(sid)
	if !c.Verify(sid, token) {
		t.Error("token válido deveria passar")
	}
	if c.Verify("outro-sid", token) {
		t.Error("token com session diferente não deveria passar")
	}
}
