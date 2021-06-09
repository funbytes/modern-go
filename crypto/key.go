package crypto

import (
	"math/rand"
	"time"
)

var allowedChars = []rune("QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm0987654321")
var specialChars = []rune("~!@#$^&*()_+-=|{}[]")

// Generate random password
// strong - include to password special characters
func GenerateKey(length int, strong bool) string {
	newRandom(time.Now().UTC().UnixNano())
	charsArray := make([]rune, length)
	for i := range charsArray {
		if strong && intToBool(i%next(3, length)) {
			charsArray[i] = specialChars[next(0, len(specialChars))]
		} else {
			charsArray[i] = allowedChars[next(0, len(allowedChars))]
		}
	}
	return string(charsArray)
}

// newRandom return random int64
func newRandom(seed int64) {
	rand.Seed(seed)
}

// next return random between min and max value
func next(min, max int) int {
	return min + rand.Intn(max-min)
}

// intToBool convert int to bool
func intToBool(i int) bool {
	if i == 1 {
		return true
	}
	return false
}
