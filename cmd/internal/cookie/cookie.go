package cookie

import (
	"qr_code/internal/cipher"
)

var cookieSecretKey = []byte("F0CElRkqPBfgg4gn4liPpbHyE0nN0RjD")

func EncryptCookie(data map[string]interface{}) (string, error) {
	return cipher.EncryptAES(data, cookieSecretKey)
}

func DecryptCookie(cookie string) (map[string]interface{}, error) {
	return cipher.DecryptAES(cookie, cookieSecretKey)
}
