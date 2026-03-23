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

const encryptedSharedKeyPrefix = "enc:v1:"

type sharedKeyCipher struct {
	aead cipher.AEAD
}

func newSharedKeyCipher(secret string) (*sharedKeyCipher, error) {
	if secret == "" {
		return nil, fmt.Errorf("%w: missing shared key secret", ErrInvalidArgument)
	}

	key := sha256.Sum256([]byte("fluent-manager:aggregation-group:" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &sharedKeyCipher{aead: aead}, nil
}

func (c *sharedKeyCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)
	return encryptedSharedKeyPrefix + base64.RawStdEncoding.EncodeToString(payload), nil
}

func (c *sharedKeyCipher) Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, encryptedSharedKeyPrefix) {
		return value, nil
	}

	raw := strings.TrimPrefix(value, encryptedSharedKeyPrefix)
	payload, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}

	nonceSize := c.aead.NonceSize()
	if len(payload) < nonceSize {
		return "", errors.New("invalid encrypted shared key payload")
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func isEncryptedSharedKey(value string) bool {
	return strings.HasPrefix(value, encryptedSharedKeyPrefix)
}
