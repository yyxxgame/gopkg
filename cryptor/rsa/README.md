# RSA - RSA 加密工具

提供 RSA 公钥/私钥加密和解密功能。

## 功能

- **公钥加密**: 使用公钥加密数据
- **公钥解密**: 使用公钥解密数据（签名验证）
- **私钥加密**: 使用私钥加密数据（数字签名）
- **私钥解密**: 使用私钥解密数据

**加密模式**: RSA/ECB/PKCS1Padding

## 使用示例

### 生成密钥对

首先需要生成 RSA 密钥对（可使用 OpenSSL 或其他工具）：

```bash
# 生成私钥
openssl genrsa -out private.pem 2048

# 从私钥提取公钥
openssl rsa -in private.pem -pubout -out public.pem

# 查看密钥内容
cat private.pem  # 私钥
cat public.pem   # 公钥
```

### 基本加密/解密

```go
package main

import (
    "encoding/base64"
    "fmt"
    "github.com/yyxxgame/gopkg/cryptor/rsa"
)

func main() {
    // 公钥和私钥字符串（从 PEM 文件读取）
    publicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----`

    privateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----`

    // 原始数据
    data := []byte("Hello, RSA!")

    // 公钥加密
    encrypted, err := rsa.PubKeyEncrypt(data, publicKey)
    if err != nil {
        panic(err)
    }
    fmt.Printf("公钥加密后：%s\n", base64.StdEncoding.EncodeToString(encrypted))

    // 私钥解密
    decrypted, err := rsa.PriKeyDecrypt(encrypted, privateKey)
    if err != nil {
        panic(err)
    }
    fmt.Printf("私钥解密后：%s\n", string(decrypted))

    // 私钥加密（数字签名）
    signed, err := rsa.PriKeyEncrypt(data, privateKey)
    if err != nil {
        panic(err)
    }
    fmt.Printf("私钥加密后：%s\n", base64.StdEncoding.EncodeToString(signed))

    // 公钥解密（验证签名）
    verified, err := rsa.PubKeyDecrypt(signed, publicKey)
    if err != nil {
        panic(err)
    }
    fmt.Printf("公钥解密后：%s\n", string(verified))
}
```

## API 参考

### 函数

#### PubKeyEncrypt

RSA 公钥加密。

```go
func PubKeyEncrypt(input []byte, pubKeyStr string) ([]byte, error)
```

**参数**:
- `input`: 要加密的数据
- `pubKeyStr`: PEM 格式的公钥字符串

#### PubKeyDecrypt

RSA 公钥解密（用于验证签名）。

```go
func PubKeyDecrypt(input []byte, pubKeyStr string) ([]byte, error)
```

**参数**:
- `input`: 要解密的数据
- `pubKeyStr`: PEM 格式的公钥字符串

#### PriKeyEncrypt

RSA 私钥加密（用于数字签名）。

```go
func PriKeyEncrypt(input []byte, priKeyStr string) ([]byte, error)
```

**参数**:
- `input`: 要加密的数据
- `priKeyStr`: PEM 格式的私钥字符串

#### PriKeyDecrypt

RSA 私钥解密。

```go
func PriKeyDecrypt(input []byte, priKeyStr string) ([]byte, error)
```

**参数**:
- `input`: 要解密的数据
- `priKeyStr`: PEM 格式的私钥字符串

## 测试

```bash
go test -v ./cryptor/rsa
```

## 典型应用场景

### 1. 敏感数据传输

```
发送方：使用接收方的公钥加密数据
接收方：使用自己的私钥解密数据
```

### 2. 数字签名

```
签名方：使用自己的私钥加密数据摘要
验证方：使用签名方的公钥解密并验证
```

### 3. 密钥交换

```
使用 RSA 加密对称密钥（如 AES 密钥）
然后使用对称加密处理大量数据
```

## 注意事项

1. **数据长度限制**: RSA 加密的数据长度受密钥长度限制
   - 2048 位密钥最多加密约 245 字节
   - 长数据应使用混合加密（RSA + AES）

2. **密钥格式**: 必须是 PEM 格式

3. **密钥安全**: 
   - 私钥必须安全存储
   - 不要将私钥提交到版本控制

4. **性能考虑**: RSA 加解密较慢，不适合大量数据

## 混合加密示例

```go
// 1. 生成 AES 密钥
aesKey := generateRandomKey(32) // 32 字节 AES-256

// 2. 使用 RSA 加密 AES 密钥
encryptedKey, _ := rsa.PubKeyEncrypt(aesKey, publicKey)

// 3. 使用 AES 加密数据
encryptedData, _ := aes.GcmEncrypt(data, aesKey)

// 4. 传输 encryptedKey 和 encryptedData
```
