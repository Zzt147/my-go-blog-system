package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var SecretKey = []byte("your-secret-key-llp-blog") // 密钥，随便写

// GenerateToken 生成 JWT Token
func GenerateToken(userId int, username string) (string, error) {
	claims := jwt.MapClaims{
		"userId":   userId,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7天过期
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}