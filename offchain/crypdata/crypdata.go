
package crypdata

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"crypto/sha256"

	"encoding/hex"

)

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



// Encrypt data
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



func Hash(data string) string {
	hashedBytes := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hashedBytes[:])
}