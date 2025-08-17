package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// EncodeBase64 Base64编码
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 Base64解码
func DecodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// EncodeBase64URL URL安全的Base64编码
func EncodeBase64URL(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeBase64URL URL安全的Base64解码
func DecodeBase64URL(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(s)
}

// EncodeHex 十六进制编码
func EncodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeHex 十六进制解码
func DecodeHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// SHA1Hash 计算SHA1哈希值
func SHA1Hash(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA256Hash 计算SHA256哈希值
func SHA256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// AESEncrypt AES加密
func AESEncrypt(plaintext, key []byte) ([]byte, error) {
	// 确保key长度为16、24或32字节
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("key length must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密并附加nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESDecrypt AES解密
func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	// 确保key长度为16、24或32字节
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("key length must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// AESEncryptString 加密字符串并返回Base64编码的结果
func AESEncryptString(plaintext, key string) (string, error) {
	encrypted, err := AESEncrypt([]byte(plaintext), []byte(key))
	if err != nil {
		return "", err
	}
	return EncodeBase64(encrypted), nil
}

// AESDecryptString 解密Base64编码的字符串
func AESDecryptString(ciphertext, key string) (string, error) {
	data, err := DecodeBase64(ciphertext)
	if err != nil {
		return "", err
	}

	decrypted, err := AESDecrypt(data, []byte(key))
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// GenerateRandomBytes 生成指定长度的随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// GenerateRandomString 生成指定长度的随机字符串（Base64编码）
func GenerateRandomString(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return EncodeBase64(bytes), nil
}

// GenerateAESKey 生成AES密钥
func GenerateAESKey(keySize int) ([]byte, error) {
	if keySize != 16 && keySize != 24 && keySize != 32 {
		return nil, errors.New("key size must be 16, 24, or 32 bytes")
	}
	return GenerateRandomBytes(keySize)
}

// SimpleXOR 简单的XOR加密/解密
func SimpleXOR(data, key []byte) []byte {
	if len(key) == 0 {
		return data
	}

	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return result
}

// SimpleXORString XOR加密字符串并返回十六进制编码
func SimpleXORString(text, key string) string {
	encrypted := SimpleXOR([]byte(text), []byte(key))
	return EncodeHex(encrypted)
}

// SimpleXORDecryptString 解密十六进制编码的XOR加密字符串
func SimpleXORDecryptString(encryptedHex, key string) (string, error) {
	encrypted, err := DecodeHex(encryptedHex)
	if err != nil {
		return "", err
	}

	decrypted := SimpleXOR(encrypted, []byte(key))
	return string(decrypted), nil
}
