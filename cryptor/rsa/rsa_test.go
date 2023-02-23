//@File     rsa_test.go
//@Time     2023/02/22
//@Author   #Suyghur,

package rsa

import (
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stringx"
	"testing"
)

const publicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCwZZwv13aFV7IduL5M0jyhTy0c
NCi5ltNNugXvqjMr1aKZlmKM7Szh6DTyhcCejSRm3RXE326Kd78nODahSw/cTKWs
cIOLnOfaqUO6l1LPUKJMXBlvdd6x6TEc6vGHXuUNonps82zzjngQNuOLIvP5Xd/c
E5+gvux3FwAE6F1/FwIDAQAB
-----END PUBLIC KEY-----`

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCwZZwv13aFV7IduL5M0jyhTy0cNCi5ltNNugXvqjMr1aKZlmKM
7Szh6DTyhcCejSRm3RXE326Kd78nODahSw/cTKWscIOLnOfaqUO6l1LPUKJMXBlv
dd6x6TEc6vGHXuUNonps82zzjngQNuOLIvP5Xd/cE5+gvux3FwAE6F1/FwIDAQAB
AoGAVxXjFQ93mP8hlWDWupB0lGLFb44kqPNYYKA9PAQ3/SUcgFzpTI/vP5xnP3Bf
FyexWANAMxDnvv69ZXxwJBMtwY/0PwdIjo+p8REtPqmKqIquuu41qFc1TvRHephv
79NCVcw89vV/bflqjJFpHnihGq+0R/i+UBg+G+Vv9CFIVdkCQQDhv3HyyihKkq/+
qwSot47vhHr+rJqbSkG3F5BvMCtYRD4MFqkdJq/v+a+Qi6ITl5qDeFfqKpBF2IfC
voPn+/YbAkEAyAkfHFHv10kkaCm8LzUMcuCUjVQ4V9tar4VcmWwEIqbpJ704EmNy
1yovBk4p2OeDK/jDdaRPPHn+WaEkc0latQJAJMHWXPD7tIDD9VSFUq9or6lDmZoj
Jmvl3VkR5HjUZe/epns+GAgHl6xxILkLr+L8frGmpvM9QJIsMNJyieBlxwJALlSG
gx2rKjbDmuiHsHtd7cF8RpuKDTc98sc1okc1Uf1MpSqbMQ8dix43FAPIh3dflzCf
vMCYpY4vzfyXn6gOvQJBAJEpUB2oPHjgrGN/nnsBUZnGl2vsx+rbJ8rtytQ51eei
jeofMKPxZSZkcxB1HtJEWu9oZOUrFDegVnQeGoXEq1U=
-----END RSA PRIVATE KEY-----`

func TestPubKeyEncryptAndPriKeyDecrypt(t *testing.T) {
	original := []byte(stringx.Rand())
	enc, err := PubKeyEncrypt(original, publicKey)
	if err != nil {
		t.Error(err)
	}
	raw, err := PriKeyDecrypt(enc, privateKey)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, raw, original)
}

func TestPriKeyEncryptAndPubKeyDecrypt(t *testing.T) {
	original := []byte(stringx.Rand())
	enc, err := PriKeyEncrypt(original, privateKey)
	if err != nil {
		t.Error(err)
	}
	raw, err := PubKeyDecrypt(enc, publicKey)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, raw, original)
}
