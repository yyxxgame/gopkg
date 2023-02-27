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
	enc, err := AesCbcPkcs7paddingEncrypt(data, aesKey, iv)
	if err != nil {
		t.Error(err)
	}
	raw, err := AesCbcPkcs7paddingDecrypt(enc, aesKey, iv)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, raw, data)
}
