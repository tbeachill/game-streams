/*
encrypt.go contains functions to encrypt and decrypt the database using AES256 encryption.
*/
package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"

	"gamestreams/config"
	"gamestreams/logs"
)

// Encrypt encrypts the database using AES256 encryption.
func Encrypt() error {
	logs.LogInfo("BCKUP", "encrypting database...", false)
	dbFile, err := os.ReadFile(config.Values.Files.Database)
	if err != nil {
		return err
	}
	key, err := os.ReadFile(config.Values.Files.EncryptionKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	dbEncrypted := gcm.Seal(nonce, nonce, dbFile, nil)
	err = os.WriteFile(config.Values.Files.EncryptedDatabase, dbEncrypted, 0777)
	if err != nil {
		return err
	}
	logs.LogInfo("BCKUP", "database encrypted", false)
	return nil
}

// Decrypt decrypts the database and writes it to the database file location.
func Decrypt() error {
	logs.LogInfo("RESTO", "decrypting database...", false)
	dbEncrypted, err := os.ReadFile(config.Values.Files.EncryptedDatabase)
	if err != nil {
		return err
	}
	key, err := os.ReadFile(config.Values.Files.EncryptionKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := dbEncrypted[:nonceSize], dbEncrypted[nonceSize:]
	dbFile, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	err = os.WriteFile(config.Values.Files.Database, dbFile, 0777)
	if err != nil {
		return err
	}
	logs.LogInfo("RESTO", "database decrypted", false)
	return nil
}
