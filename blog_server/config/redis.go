package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Ctx = context.Background()

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 你的 Redis 地址
		Password: "",               // 你的 Redis 密码 (如果有)
		DB:       0,                // 默认 DB 0
	})

	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		fmt.Println("❌ Redis 连接失败:", err)
	} else {
		fmt.Println("✅ Redis 连接成功！")
	}
}