//@File     rsa.go
//@Time     2023/02/22
//@Author   #Suyghur,

package cryptor

import (
	"bytes"
	"io"
)

// PubKeyEncrypt RSA/ECB/PKCS1Padding公钥加密
func PubKeyEncrypt(input []byte, pubKeyStr string) ([]byte, error) {
	pubKey, err := getPubKey([]byte(pubKeyStr))
	if err != nil {
		return []byte(""), err
	}
	output := bytes.NewBuffer(nil)
	if err = pubKeyIO(pubKey, bytes.NewReader(input), output, true); err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// PubKeyDecrypt RSA/ECB/PKCS1Padding公钥解密
func PubKeyDecrypt(input []byte, pubKeyStr string) ([]byte, error) {
	pubKey, err := getPubKey([]byte(pubKeyStr))
	if err != nil {
		return []byte(""), err
	}
	output := bytes.NewBuffer(nil)
	if err = pubKeyIO(pubKey, bytes.NewReader(input), output, false); err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// PriKeyEncrypt RSA/ECB/PKCS1Padding私钥加密
func PriKeyEncrypt(input []byte, priKeyStr string) ([]byte, error) {
	priKey, err := getPriKey([]byte(priKeyStr))
	if err != nil {
		return []byte(""), err
	}
	output := bytes.NewBuffer(nil)
	if err = priKeyIO(priKey, bytes.NewReader(input), output, true); err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// PriKeyDecrypt RSA/ECB/PKCS1Padding私钥解密
func PriKeyDecrypt(input []byte, priKeyStr string) ([]byte, error) {
	priKey, err := getPriKey([]byte(priKeyStr))
	if err != nil {
		return []byte(""), err
	}
	output := bytes.NewBuffer(nil)
	if err = priKeyIO(priKey, bytes.NewReader(input), output, false); err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}
