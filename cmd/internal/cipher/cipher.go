package cipher

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

// md5 от строки
func MD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

func EncodeBase64(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// Функция для декодирования base64 строки
func DecodeBase64(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
