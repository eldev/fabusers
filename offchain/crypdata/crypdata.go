/*
This package is used to encrypt and decrypt private data.
Also, it provides hash (sha256) procedure for passwords.
*/
package crypdata

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// At the first stage, we can use the single pair
// of private and public keys to encrypt and decrypt private data
// Idealy, we should use users public and private keys in the Fabric
var privKey *rsa.PrivateKey = nil

func Init() error {
	// Generate RSA Keys
	privkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.New("can't generate rsa keys")
	}
	privKey = privkey
	return nil
}

// Encrypt() encrypts data and returns ciphertext
func Encrypt(data []byte) ([]byte, error) {

	if privKey == nil {
		return nil, errors.New("No rsa keys, init crypdata package")
	}
	label := []byte("")

	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &privKey.PublicKey, data, label)
	if err != nil {
		return nil, errors.New("Encryption error")
	} else {
		return ciphertext, nil
	}
}

// Decrypt() decrypts ciphertext
func Decrypt(ciphertext []byte) ([]byte, error) {

	if privKey == nil {
		return nil, errors.New("No rsa keys, init crypdata package")
	}

	label := []byte("")
	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader
	//
	plainText, err := rsa.DecryptOAEP(sha256.New(), rng, privKey, ciphertext, label)

	if err != nil {
		return nil, errors.New("Decryption error")
	} else {
		return plainText, nil
	}
}

// Hash() calculates sha256 hash of data
func Hash(data string) string {
	hashedBytes := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hashedBytes[:])
}
