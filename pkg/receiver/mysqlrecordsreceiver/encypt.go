package mysqlrecordsreceiver

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"os"

	"go.uber.org/zap"
)

var randombytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func readMySecret(conf *Config) (string, error) {
	secret, err := os.ReadFile(conf.EncryptSecretPath)
	if err != nil {
		return "", err
	}
	EncryptSecret := string(secret)
	return EncryptSecret, nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
func Decode(s string, logger *zap.Logger) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		logger.Error("error during decoding password", zap.Error(err))
	}
	return data
}

// Encrypt method is to encrypt or hide any classified text
func Encrypt(text, MySecret string, logger *zap.Logger) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		logger.Error("error during encryption", zap.Error(err))
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, randombytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

// Decrypt method is to extract back the encrypted text
func Decrypt(text, MySecret string, logger *zap.Logger) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		logger.Error("error during decryption", zap.Error(err))
	}
	cipherText := Decode(text, logger)
	cfb := cipher.NewCFBDecrypter(block, randombytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}
