package config

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

// [NEW] 全局配置结构体
type AppConfig struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	Mail struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"mail"`
	File struct {
		UploadImagesDir string `yaml:"upload_images_dir"`
		UploadAvatarDir string `yaml:"upload_avatar_dir"`
		ArticleImgDir   string `yaml:"article_img_dir"`
	} `yaml:"file"`
}

var Config AppConfig

// [NEW] 初始化配置
func InitConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("❌ 读取配置文件失败: %v", err)
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Fatalf("❌ 解析配置文件失败: %v", err)
	}
	log.Println("✅ 配置文件加载成功")
}
