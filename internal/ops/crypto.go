package ops

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const encryptedMagic = "SNAPDBENC1"

func encryptFile(path, passphrase string) (string, error) {
	if passphrase == "" {
		return "", errors.New("encryption passphrase is required")
	}

	plaintext, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	key, err := scrypt.Key([]byte(passphrase), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, []byte(encryptedMagic))
	out := append([]byte(encryptedMagic), salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	outPath := path + ".enc"
	if err := os.WriteFile(outPath, out, 0600); err != nil {
		return "", err
	}
	return outPath, os.Remove(path)
}

func decryptFile(path, passphrase string) (string, error) {
	if passphrase == "" {
		return "", errors.New("decryption passphrase is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	headerLen := len(encryptedMagic) + 16
	if len(data) <= headerLen {
		return "", fmt.Errorf("%s is not a SnapDB encrypted file", path)
	}
	if string(data[:len(encryptedMagic)]) != encryptedMagic {
		return "", fmt.Errorf("%s is not a SnapDB encrypted file", path)
	}

	salt := data[len(encryptedMagic):headerLen]
	key, err := scrypt.Key([]byte(passphrase), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceStart := headerLen
	nonceEnd := nonceStart + gcm.NonceSize()
	if len(data) <= nonceEnd {
		return "", fmt.Errorf("%s is missing encrypted payload", path)
	}

	plaintext, err := gcm.Open(nil, data[nonceStart:nonceEnd], data[nonceEnd:], []byte(encryptedMagic))
	if err != nil {
		return "", err
	}

	outPath := strings.TrimSuffix(path, ".enc")
	if outPath == path {
		outPath = path + ".dec"
	}
	if err := os.WriteFile(outPath, plaintext, 0600); err != nil {
		return "", err
	}
	return outPath, nil
}
