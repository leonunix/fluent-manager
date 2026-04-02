package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const encryptedBootstrapSecretPrefix = "enc:bootstrap:v1:"
const encryptedAgentKeyPrefix = "enc:agentkey:v1:"

type bootstrapSecretCipher struct {
	aead   cipher.AEAD
	prefix string
}

func newBootstrapSecretCipher(secret string) (*bootstrapSecretCipher, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("%w: missing bootstrap secret", ErrInvalidArgument)
	}

	key := sha256.Sum256([]byte("fluent-manager:bootstrap-host:" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &bootstrapSecretCipher{aead: aead, prefix: encryptedBootstrapSecretPrefix}, nil
}

func newAgentAccessKeyCipher(secret string) (*bootstrapSecretCipher, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("%w: missing secret", ErrInvalidArgument)
	}
	key := sha256.Sum256([]byte("fluent-manager:agent-access-key:" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &bootstrapSecretCipher{aead: aead, prefix: encryptedAgentKeyPrefix}, nil
}

func (c *bootstrapSecretCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)
	return c.prefix + base64.RawStdEncoding.EncodeToString(payload), nil
}

func (c *bootstrapSecretCipher) Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, c.prefix) {
		return value, nil
	}

	raw := strings.TrimPrefix(value, c.prefix)
	payload, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}

	nonceSize := c.aead.NonceSize()
	if len(payload) < nonceSize {
		return "", errors.New("invalid encrypted bootstrap secret payload")
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
