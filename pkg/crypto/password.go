// Package crypto 提供加密相关工具
package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost bcrypt 默认加密强度
	DefaultCost = 12
	// 密码字符集
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberBytes   = "0123456789"
	specialBytes  = "!@#$%^&*"
	allCharBytes  = letterBytes + numberBytes + specialBytes
)

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}
	return string(hash), nil
}

// CheckPassword 验证密码是否正确
func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GeneratePassword 生成随机密码
// length: 密码长度
// 返回包含字母、数字和特殊字符的随机密码
func GeneratePassword(length int) (string, error) {
	if length < 8 {
		length = 12
	}

	password := make([]byte, length)

	// 确保至少包含一个字母、数字和特殊字符
	if length >= 8 {
		// 至少一个字母
		char, err := randomChar(letterBytes)
		if err != nil {
			return "", err
		}
		password[0] = char

		// 至少一个数字
		char, err = randomChar(numberBytes)
		if err != nil {
			return "", err
		}
		password[1] = char

		// 至少一个特殊字符
		char, err = randomChar(specialBytes)
		if err != nil {
			return "", err
		}
		password[2] = char

		// 其余随机字符
		for i := 3; i < length; i++ {
			char, err := randomChar(allCharBytes)
			if err != nil {
				return "", err
			}
			password[i] = char
		}

		// 打乱顺序
		if err := shuffle(password); err != nil {
			return "", err
		}
	} else {
		// 长度不足8时，全部使用随机字符
		for i := 0; i < length; i++ {
			char, err := randomChar(allCharBytes)
			if err != nil {
				return "", err
			}
			password[i] = char
		}
	}

	return string(password), nil
}

// randomChar 从字符集中随机选择一个字符
func randomChar(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, fmt.Errorf("generate random char: %w", err)
	}
	return charset[n.Int64()], nil
}

// shuffle 打乱字节数组
func shuffle(data []byte) error {
	for i := len(data) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return fmt.Errorf("shuffle data: %w", err)
		}
		j := n.Int64()
		data[i], data[j] = data[j], data[i]
	}
	return nil
}
