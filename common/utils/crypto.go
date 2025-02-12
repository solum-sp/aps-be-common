package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type cryptoType struct{}

var Crypto cryptoType

var _IV = []byte("!!SolumVina@2025")

func (c *cryptoType) EncryptString(key string, content string) string {

	k1 := []byte(key)
	data := []byte(content)
	block, _ := aes.NewCipher(k1)
	stream := cipher.NewCFBEncrypter(block, _IV)
	stream.XORKeyStream(data, data)
	return fmt.Sprintf("%x", data)
}

func (c *cryptoType) DecryptString(key string, content string) string {

	bytes, _ := hex.DecodeString(content)
	block, _ := aes.NewCipher([]byte(key))
	stream := cipher.NewCFBDecrypter(block, _IV)

	stream.XORKeyStream(bytes, bytes)

	return string(bytes)
}

func (c *cryptoType) Base64String(content string) string {
	rawDecodedText := base64.StdEncoding.EncodeToString([]byte(content))
	return rawDecodedText
}

func (c *cryptoType) DecodeBase64String(content string) string {
	rawDecodedText, _ := base64.StdEncoding.DecodeString(content)
	return string(rawDecodedText)
}

func (c *cryptoType) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (c *cryptoType) ComparePassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
