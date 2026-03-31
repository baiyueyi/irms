package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

type credentialCrypto struct {
	key string
}

func newCredentialCrypto(key string) *credentialCrypto {
	return &credentialCrypto{key: key}
}

func (c *credentialCrypto) encryptPlaintext(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key32())
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)
	buf := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(buf), nil
}

func (c *credentialCrypto) decryptCiphertext(ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(c.key32())
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < aead.NonceSize() {
		return "", errors.New("invalid ciphertext")
	}
	nonce := raw[:aead.NonceSize()]
	payload := raw[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (c *credentialCrypto) key32() []byte {
	sum := sha256.Sum256([]byte(c.key))
	return sum[:]
}
