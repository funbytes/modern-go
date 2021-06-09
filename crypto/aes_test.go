package crypto

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAesCBCWithBase64(t *testing.T) {
	Convey("Test encrypt decrypt", t, func(c C) {
		key := []byte("1234567890123456")
		s, err := AesEncryptCBCWithBase64([]byte("1589598989"), key)
		fmt.Print(s)
		c.So(err, ShouldBeNil)
		s, err = AesDecryptCBCWithBase64([]byte(s), key)
		c.So(err, ShouldBeNil)
		c.So(s, ShouldEqual, "1589598989")
	})

	Convey("Exception", t, func(c C) {
		key := []byte("1234567890123456")
		_, err := AesDecryptCBCWithBase64([]byte("wqewq"), key)
		c.So(err, ShouldNotBeNil)
	})

	Convey("Exception", t, func(c C) {
		key := []byte("1234567890123456")
		_, err := AesDecryptCBCWithBase64([]byte(""), key)
		c.So(err, ShouldBeNil)
	})

	Convey("Exception", t, func(c C) {
		key := []byte("1234567890123456")
		_, err := AesDecryptCBCWithBase64([]byte("2332"), key)
		c.So(err, ShouldNotBeNil)
	})
}
