package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

type payload struct {
	Subject string `json:"sub"`
	Issuer  string `json:"iss"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

func NewManager(secret string, issuer string, ttl time.Duration) *Manager {
	return &Manager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

func (m *Manager) Generate(subject string) (string, error) {
	now := time.Now().UTC()
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	body := payload{
		Subject: subject,
		Issuer:  m.issuer,
		Issued:  now.Unix(),
		Expires: now.Add(m.ttl).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	unsigned := encode(headerJSON) + "." + encode(bodyJSON)
	signature := m.sign(unsigned)
	return unsigned + "." + signature, nil
}

func (m *Manager) Validate(raw string) (string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return "", ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(m.sign(unsigned))) {
		return "", ErrInvalidToken
	}

	bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ErrInvalidToken
	}

	var body payload
	if err := json.Unmarshal(bodyJSON, &body); err != nil {
		return "", ErrInvalidToken
	}

	if body.Issuer != m.issuer || body.Subject == "" {
		return "", ErrInvalidToken
	}
	if time.Now().UTC().Unix() > body.Expires {
		return "", ErrExpiredToken
	}

	return body.Subject, nil
}

func encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func (m *Manager) sign(unsigned string) string {
	sum := hmac.New(sha256.New, m.secret)
	_, _ = sum.Write([]byte(unsigned))
	return fmt.Sprintf("%s", encode(sum.Sum(nil)))
}
