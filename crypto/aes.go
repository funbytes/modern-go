package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

func AesEncryptCBCWithBase64(origData []byte, key []byte) (string, error) {
	bts, err := AesEncryptCBC(key, origData)
	return base64.StdEncoding.EncodeToString(bts), err
}

func AesEncryptCBC(key []byte, origData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	bts := make([]byte, len(origData))
	blockMode.CryptBlocks(bts, origData)
	return bts, nil
}

func AesDecryptCBCWithBase64(encrypted []byte, key []byte) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(string(encrypted))
	if err != nil {
		return "", err
	}
	decrypted, err := AesDecryptCBC(key, encrypted)
	return string(decrypted), err
}

func AesDecryptCBC(key []byte, encrypted []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(encrypted)%blockSize != 0 {
		return nil, errors.New("encrypted len invalid")
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	decrypted := make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted, err = pkcs5UnPadding(decrypted)
	return decrypted, err
}

func pkcs5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return origData, nil
	}
	unPadding := int(origData[length-1])
	if length-unPadding < 0 {
		return nil, errors.New("un padding failed")
	}
	return origData[:(length - unPadding)], nil
}
