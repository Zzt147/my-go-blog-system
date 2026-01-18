package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	// [NEW] 确保配置已加载
	if Config.Database.DSN == "" {
		InitConfig()
	}

	// [MODIFY] 使用配置文件中的 DSN
	dsn := Config.Database.DSN

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("✅ MySQL 数据库连接成功！")
}
