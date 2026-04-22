package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SessionManager struct {
	secret []byte
	ttl    time.Duration
}

func NewSessionManager(secret string, ttl time.Duration) *SessionManager {
	return &SessionManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *SessionManager) Sign(user string) (string, error) {
	expiry := time.Now().Add(m.ttl).Unix()
	payload := fmt.Sprintf("%s.%d", user, expiry)
	mac := m.computeMAC(payload)
	return payload + "." + mac, nil
}

func (m *SessionManager) Verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("formato inválido")
	}
	user, expiryStr, sigGiven := parts[0], parts[1], parts[2]
	payload := user + "." + expiryStr

	sigExpected := m.computeMAC(payload)
	if !hmac.Equal([]byte(sigGiven), []byte(sigExpected)) {
		return "", fmt.Errorf("mac inválido")
	}
	expiry, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("expiry inválida")
	}
	if time.Now().Unix() > expiry {
		return "", fmt.Errorf("expirado")
	}
	return user, nil
}

func (m *SessionManager) computeMAC(payload string) string {
	h := hmac.New(sha256.New, m.secret)
	h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
