package outreachaccounts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type credentialPayload struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

type credentialCipher struct {
	aead cipher.AEAD
}

func newCredentialCipher(encodedKey string) (*credentialCipher, error) {
	encodedKey = strings.TrimSpace(encodedKey)
	if encodedKey == "" {
		return nil, ErrEncryptionUnavailable
	}
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("%w: key must be standard base64 encoding of exactly 32 bytes", ErrEncryptionUnavailable)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("initialize outreach credential encryption: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize outreach credential encryption mode: %w", err)
	}
	return &credentialCipher{aead: aead}, nil
}

func (vault *credentialCipher) encrypt(accountKey, mailbox string, payload credentialPayload) ([]byte, error) {
	if vault == nil || vault.aead == nil {
		return nil, ErrEncryptionUnavailable
	}
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode outreach credentials: %w", err)
	}
	nonce := make([]byte, vault.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate outreach credential nonce: %w", err)
	}
	sealed := vault.aead.Seal(nil, nonce, plaintext, credentialAAD(accountKey, mailbox))
	return append(nonce, sealed...), nil
}

func (vault *credentialCipher) decrypt(accountKey, mailbox string, encoded []byte) (credentialPayload, error) {
	if vault == nil || vault.aead == nil {
		return credentialPayload{}, ErrEncryptionUnavailable
	}
	nonceSize := vault.aead.NonceSize()
	if len(encoded) <= nonceSize {
		return credentialPayload{}, fmt.Errorf("encrypted outreach credential payload is invalid")
	}
	plaintext, err := vault.aead.Open(nil, encoded[:nonceSize], encoded[nonceSize:], credentialAAD(accountKey, mailbox))
	if err != nil {
		return credentialPayload{}, fmt.Errorf("decrypt outreach credential payload: authentication failed")
	}
	var payload credentialPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return credentialPayload{}, fmt.Errorf("decode outreach credential payload: invalid payload")
	}
	if err := validateCredentialPayload(payload); err != nil {
		return credentialPayload{}, err
	}
	return payload, nil
}

func credentialAAD(accountKey, mailbox string) []byte {
	return []byte(strings.TrimSpace(accountKey) + "\x00" + strings.ToLower(strings.TrimSpace(mailbox)))
}
