package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// [NEW] 解析 Token
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
