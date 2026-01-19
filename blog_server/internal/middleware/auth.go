package middleware

import (
	"my-blog/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// [NEW] JWT 认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization Header
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			// 尝试从 query 中获取 (可选兼容)
			tokenStr = c.Query("token")
		}

		// 2. 简单处理 Bearer 前缀 (如果有)
		if strings.HasPrefix(tokenStr, "Bearer ") {
			tokenStr = tokenStr[7:]
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, utils.Error("请先登录"))
			c.Abort()
			return
		}

		// 3. 解析 Token
		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.Error("登录已过期，请重新登录"))
			c.Abort()
			return
		}

		// 4. 将用户信息存入 Context
		// 注意：JWT 解析出的数字默认是 float64
		if userIdFloat, ok := claims["userId"].(float64); ok {
			c.Set("userId", int(userIdFloat))
		}
		if username, ok := claims["username"].(string); ok {
			c.Set("username", username)
		}

		c.Next()
	}
}
