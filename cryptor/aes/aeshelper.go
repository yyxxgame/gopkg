package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

func CbcPkcs7Encrypter(data, key []byte) ([]byte, error) {
	if block, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		blockSize := block.BlockSize()
		data = PKCS7Padding(data, blockSize)
		enc := make([]byte, len(data))
		blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
		blockMode.CryptBlocks(enc, data)
		return enc, nil
	}
}

func CbcPkcs7Decrypter(data, key []byte) ([]byte, error) {
	if block, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		blockSize := block.BlockSize()
		blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
		if len(data)%blockSize != 0 {
			return nil, errors.New("input not full blocks")
		}
		raw := make([]byte, len(data))
		blockMode.CryptBlocks(raw, data)
		raw = PKCS7UnPadding(raw)
		return raw, nil
	}
}

func EncryptCbcPkcs7(encodeToStringFunc func([]byte) string, data string, key string) (string, error) {
	if enc, err := CbcPkcs7Encrypter([]byte(data), []byte(key)); err != nil {
		return "", err
	} else {
		return encodeToStringFunc(enc), nil
	}
}

func DecryptCbcPkcs7(decodeStringFunc func(string) ([]byte, error), data string, key string) (string, error) {
	if raw, err := decodeStringFunc(data); err != nil {
		return "", err
	} else {
		res, err := CbcPkcs7Decrypter(raw, []byte(key))
		return string(res), err
	}
}
