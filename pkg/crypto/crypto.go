// Package crypto implement functions for string crypt
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
)

// encKey rand key
type encData struct {
	aesGCM cipher.AEAD
	nonce  []byte
}

// Instance crypto entity
var instance *encData

// Encode string by GCM algorithm and get hex
func Encode(str string) (string, error) {
	// Init encrypt data
	if err := keyInit(); err != nil {
		return "", err
	}
	src := []byte(str)
	// Encrypt userId
	dst := instance.aesGCM.Seal(nil, instance.nonce, src, nil)
	// Get hexadecimal string from encode string
	sha := hex.EncodeToString(dst)

	return sha, nil
}

// Decode shaStr from encrypted string
func Decode(shaStr string) (string, error) {
	// Init encrypt data
	if err := keyInit(); err != nil {
		return "", err
	}
	// Convert to bytes from hex
	dst, err := hex.DecodeString(shaStr)
	if err != nil {
		return "", err
	}
	// Decode
	src, err := instance.aesGCM.Open(nil, instance.nonce, dst, nil)
	if err != nil {
		return "", err
	}

	return string(src), nil
}

// keyInit init crypt params
func keyInit() error {
	// If you need generate new key
	if instance == nil {
		key, err := generateRandom(aes.BlockSize)
		if err != nil {
			return err
		}

		aesBlock, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		aesGCM, err := cipher.NewGCM(aesBlock)
		if err != nil {
			return err
		}
		// initialize vector
		nonce, err := generateRandom(aesGCM.NonceSize())
		if err != nil {
			return err
		}
		// Allocation enc type
		instance = new(encData)
		instance.aesGCM = aesGCM
		instance.nonce = nonce
	}

	return nil
}

// generateRandom byte slice
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
