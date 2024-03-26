//@File     aes_test.go
//@Time     2023/02/23
//@Author   #Suyghur,

package aes

import (
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stringx"
	"testing"
)

func TestAESCbcEncryptAndDecrypt(t *testing.T) {
	aesKey := []byte(stringx.Randn(16))
	iv := GenerateIV(aesKey)
	data := []byte(stringx.Randn(200))

	enc, err := CbcPkcs7paddingEncrypt(data, aesKey, iv)
	assert.Nil(t, err)

	raw, err := CbcPkcs7paddingDecrypt(enc, aesKey, iv)
	assert.Nil(t, err)

	assert.Equal(t, raw, data)
}

func TestAESGcmEndcryptAndDecrypt(t *testing.T) {
	key := []byte(stringx.Randn(16))
	data := []byte(stringx.Randn(200))

	enc, err := GcmEncrypt(data, key)
	assert.Nil(t, err)

	raw, err := GcmDecrypt(enc, key)
	assert.Nil(t, err)

	assert.Equal(t, raw, data)
}
