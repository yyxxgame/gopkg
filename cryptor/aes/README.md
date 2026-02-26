# AES - AES 加密工具

提供 AES 加密/解密功能，支持多种加密模式。

## 支持的加密模式

- **AES/CBC/PKCS7Padding**: CBC 模式，PKCS7 填充
- **AES/GCM**: GCM 模式（认证加密）

## 使用示例

### CBC 模式加密/解密

```go
package main

import (
    "encoding/base64"
    "fmt"
    "github.com/yyxxgame/gopkg/cryptor/aes"
)

func main() {
    // 密钥（必须是 16、24 或 32 字节）
    key := []byte("0123456789abcdef") // 16 字节 = AES-128
    
    // 生成 IV
    iv := aes.GenerateIV(key)
    
    // 原始数据
    data := []byte("Hello, World!")
    
    // CBC 加密
    encrypted, err := aes.CbcPkcs7paddingEncrypt(data, key, iv)
    if err != nil {
        panic(err)
    }
    fmt.Printf("加密后：%s\n", base64.StdEncoding.EncodeToString(encrypted))
    
    // CBC 解密
    decrypted, err := aes.CbcPkcs7paddingDecrypt(encrypted, key, iv)
    if err != nil {
        panic(err)
    }
    fmt.Printf("解密后：%s\n", string(decrypted))
}
```

### GCM 模式加密/解密

```go
package main

import (
    "encoding/base64"
    "fmt"
    "github.com/yyxxgame/gopkg/cryptor/aes"
)

func main() {
    // 密钥（必须是 16、24 或 32 字节）
    key := []byte("0123456789abcdef") // 16 字节 = AES-128
    
    // 原始数据
    data := []byte("Hello, World!")
    
    // GCM 加密
    encrypted, err := aes.GcmEncrypt(data, key)
    if err != nil {
        panic(err)
    }
    fmt.Printf("加密后：%s\n", base64.StdEncoding.EncodeToString(encrypted))
    
    // GCM 解密
    decrypted, err := aes.GcmDecrypt(encrypted, key)
    if err != nil {
        panic(err)
    }
    fmt.Printf("解密后：%s\n", string(decrypted))
}
```

## API 参考

### 函数

#### PKCS7Padding

PKCS7 填充。

```go
func PKCS7Padding(ciphertext []byte, blockSize int) []byte
```

#### PKCS7UnPadding

PKCS7 去填充。

```go
func PKCS7UnPadding(origData []byte) []byte
```

#### GenerateIV

生成 IV（初始向量）。

```go
func GenerateIV(key []byte) []byte
```

#### CbcPkcs7paddingEncrypt

AES/CBC/PKCS7Padding 加密。

```go
func CbcPkcs7paddingEncrypt(data, key, iv []byte) ([]byte, error)
```

**参数**:
- `data`: 原始数据
- `key`: 密钥（16/24/32 字节）
- `iv`: 初始向量（长度与密钥相同）

#### CbcPkcs7paddingDecrypt

AES/CBC/PKCS7Padding 解密。

```go
func CbcPkcs7paddingDecrypt(data, key, iv []byte) ([]byte, error)
```

#### GcmEncrypt

AES/GCM 加密。

```go
func GcmEncrypt(raw, key []byte) ([]byte, error)
```

**参数**:
- `raw`: 原始数据
- `key`: 密钥（16/24/32 字节）

**返回值**:
- 包含 nonce 和密文的组合数据

#### GcmDecrypt

AES/GCM 解密。

```go
func GcmDecrypt(enc, key []byte) ([]byte, error)
```

**参数**:
- `enc`: GcmEncrypt 返回的加密数据
- `key`: 密钥

## 密钥长度说明

AES 支持三种密钥长度：

| 密钥长度 | 字节数 | 名称 |
|---------|-------|------|
| 128 位 | 16 字节 | AES-128 |
| 192 位 | 24 字节 | AES-192 |
| 256 位 | 32 字节 | AES-256 |

## 测试

```bash
go test -v ./cryptor/aes
```

## 安全建议

1. **密钥管理**: 使用安全的密钥存储方案，不要硬编码密钥
2. **IV 使用**: CBC 模式下，每次加密应使用不同的 IV
3. **模式选择**: 
   - GCM 模式提供认证加密，推荐使用
   - CBC 模式需要额外的 MAC 验证
4. **密钥长度**: 建议使用 AES-256（32 字节）以获得更高安全性

## 应用场景

- 敏感数据加密存储
- 配置文件加密
- 通信数据加密
- 密码/凭证加密
