package ui

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type CSRFManager struct {
	secret []byte
}

func NewCSRFManager(secret string) *CSRFManager {
	return &CSRFManager{secret: []byte(secret)}
}

func (c *CSRFManager) Generate(sessionID string) string {
	h := hmac.New(sha256.New, c.secret)
	h.Write([]byte(sessionID))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (c *CSRFManager) Verify(sessionID, token string) bool {
	expected := c.Generate(sessionID)
	return hmac.Equal([]byte(expected), []byte(token))
}
