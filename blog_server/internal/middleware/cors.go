package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// Cors 处理跨域请求,支持 options 访问
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 对应 Java: allowedOriginPatterns("*")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin) // 允许所有来源
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE") // 允许的方法
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true") // 对应 allowCredentials(true)
		}

		// 允许放行 OPTIONS 请求 (预检请求)
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next() // 放行，进入下一个处理环节
	}
}