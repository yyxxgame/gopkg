//@File     aes.go
//@Time     2023/02/23
//@Author   #Suyghur,

package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func GenerateIV(key []byte) []byte {
	length := len(key)
	iv := make([]byte, length)
	for i := 0; i < length; i++ {
		iv[i] = key[length-1-i]
	}
	return iv
}

func GenerateNonce() []byte {
	nonce := make([]byte, 12)
	_, _ = io.ReadFull(rand.Reader, nonce)
	return nonce
}

// CbcPkcs7paddingEncrypt AES/CBC/PKCS7Padding加密
func CbcPkcs7paddingEncrypt(data, key, iv []byte) ([]byte, error) {
	if block, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		blockSize := block.BlockSize()
		data = PKCS7Padding(data, blockSize)
		blockMode := cipher.NewCBCEncrypter(block, iv)
		enc := make([]byte, len(data))
		blockMode.CryptBlocks(enc, data)
		return enc, nil
	}
}

// CbcPkcs7paddingDecrypt AES/CBC/PKCS7Padding解密
func CbcPkcs7paddingDecrypt(data, key, iv []byte) ([]byte, error) {
	if block, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		blockMode := cipher.NewCBCDecrypter(block, iv)
		raw := make([]byte, len(data))
		blockMode.CryptBlocks(raw, data)
		raw = PKCS7UnPadding(raw)
		return raw, nil
	}
}

// GcmEncrypt AES/GCM 加密
func GcmEncrypt(data, key, nonce []byte) ([]byte, error) {
	if len(nonce) != 12 {
		return nil, errors.New("nonce size must is 12")
	}
	if block, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		aesGcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		return aesGcm.Seal(nil, nonce, data, nil), nil
	}
}

// GcmDecrypt AES/GCM 解密
func GcmDecrypt(data, key, nonce []byte) ([]byte, error) {
	if block, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		aesGcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		return aesGcm.Open(nil, nonce, data, nil)
	}
}
