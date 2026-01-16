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
	// 1. 配置 DSN (Data Source Name)
	// 根据你的 application-jdbc.yml: 用户 root, 密码 57502, 库 my_blog_system
	username := "root"
	password := "57502"
	host := "127.0.0.1"
	port := "3306"
	dbname := "my_blog_system"
	
	// 拼接触发字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	var err error
	// 2. 连接数据库 (开启日志，方便调试 SQL)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}

	// 3. 设置连接池 (对应你的 Druid 配置)
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(5)  // min-idle
	sqlDB.SetMaxOpenConns(20) // max-active
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("✅ MySQL 数据库连接成功！")
}