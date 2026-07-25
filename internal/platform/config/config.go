package config

import (
	"log"
	"os"

	"github.com/godotenv/godotenv"
)

var Config AppConfig

type AppConfig struct {
	Port               string
	GinMode            string
	APIToken           string
	TencentCloudSecretId  string
	TencentCloudSecretKey string
	TencentCloudRegion    string
}

func Init() {
	// 尝试加载 .env 文件，文件不存在不报错
	if err := godotenv.Load(); err != nil {
		log.Printf("[config] .env 文件不存在，将使用系统环境变量")
	}

	Config = AppConfig{
		Port:               getEnv("PORT", "8000"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		APIToken:           getEnv("API_TOKEN", ""),
		TencentCloudSecretId:  getEnv("TENCENTCLOUD_SECRET_ID", ""),
		TencentCloudSecretKey: getEnv("TENCENTCLOUD_SECRET_KEY", ""),
		TencentCloudRegion:    getEnv("TENCENTCLOUD_REGION", ""),
	}

	if Config.APIToken == "" {
		log.Fatal("[config] API_TOKEN 未配置，请设置环境变量")
	}
	if Config.TencentCloudSecretId == "" || Config.TencentCloudSecretKey == "" {
		log.Fatal("[config] 腾讯云 SecretId/SecretKey 未配置")
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}