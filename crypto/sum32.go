package crypto

import (
	"fmt"
	"hash/crc32"
)

func Sum32(b []byte) (sum uint64) {
	h := crc32.NewIEEE()
	h.Write(b)
	sum = uint64(h.Sum32())
	return
}
func Sum32String(b []byte) string {
	return fmt.Sprintf("%x", Sum32(b))
}
