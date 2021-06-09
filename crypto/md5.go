package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5String(b []byte) string {
	hasher := md5.New()
	hasher.Write(b)
	return hex.EncodeToString(hasher.Sum(nil))
}
