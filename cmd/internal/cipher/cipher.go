package cipher

import (
	"crypto/md5"
	"encoding/hex"
)

// md5 от строки
func MD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}
