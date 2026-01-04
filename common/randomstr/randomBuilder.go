package randomstr

import (
	"crypto/rand"
	"errors"
	"math/big"
	mathrand "math/rand"
	"time"
	"unsafe"
)

type RandomSource int

const (
	CryptoSecure RandomSource = iota
	MathRand
	TimeBased
)

func GenerateFromCharset(length int, charset string, source RandomSource) (string, error) {
	if length <= 0 {
		return "", errors.New("长度必须大于0")
	}
	if len(charset) == 0 {
		return "", errors.New("字符集不能为空")
	}
	if length > 1000000 {
		return "", errors.New("长度过大，最大支持1000000个字符")
	}

	switch source {
	case CryptoSecure:
		return generateCryptoSecure(length, charset)
	case MathRand:
		return generateMathRand(length, charset)
	case TimeBased:
		return generateTimeBased(length, charset)
	default:
		return "", errors.New("无效的随机源")
	}
}

// generateCryptoSecure 使用加密安全随机源生成字符串
func generateCryptoSecure(length int, charset string) (string, error) {
	b := make([]byte, length)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	charCount := big.NewInt(int64(len(charset)))
	for i := range b {
		num := big.NewInt(0).SetBytes([]byte{b[i]})
		idx := int(num.Mod(num, charCount).Int64())
		b[i] = charset[idx]
	}

	return string(b), nil
}

func generateMathRand(length int, charset string) (string, error) {
	mathrand.Seed(time.Now().UnixNano())

	b := make([]byte, length)

	for i := range b {
		b[i] = charset[mathrand.Intn(len(charset))]
	}

	return string(b), nil
}

func generateTimeBased(length int, charset string) (string, error) {
	b := make([]byte, length)

	now := time.Now().UnixNano()

	for i := range b {
		seed := now + int64(i)*7919
		idx := int(seed % int64(len(charset)))
		b[i] = charset[idx]
	}

	return string(b), nil
}

func GenerateNumeric(length int, source RandomSource) (string, error) {
	return GenerateFromCharset(length, "0123456789", source)
}

func GenerateAlphanumeric(length int, source RandomSource) (string, error) {
	return GenerateFromCharset(length, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", source)
}

func GenerateHex(length int, source RandomSource) (string, error) {
	return GenerateFromCharset(length, "0123456789abcdef", source)
}

func GenerateCustom(length int, charset string, source RandomSource) (string, error) {
	return GenerateFromCharset(length, charset, source)
}

func generateNoAlloc(length int, charset string, source RandomSource) string {
	b := make([]byte, length)

	switch source {
	case CryptoSecure:
		rand.Read(b)
	case MathRand:
		mathrand.Seed(time.Now().UnixNano())
		for i := range b {
			b[i] = byte(mathrand.Intn(256))
		}
	case TimeBased:
		now := time.Now().UnixNano()
		for i := range b {
			b[i] = byte((now + int64(i)*7919) % 256)
		}
	}

	charCount := len(charset)
	for i := range b {
		idx := int(b[i]) % charCount
		b[i] = charset[idx]
	}

	return unsafe.String(unsafe.SliceData(b), len(b))
}
