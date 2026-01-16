package main

import (
	"my-blog/config"
	"my-blog/internal/router"
)

func main() {
	// 1. 初始化数据库
	config.InitDB()

	// 2. 初始化路由
	r := router.InitRouter()

	// 3. 启动端口 (你的 application.yml 写的是 8080)
	r.Run(":8080")
}