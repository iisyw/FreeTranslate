package config

import (
	"log"
	"os"

	"github.com/godotenv/godotenv"
)

var Config AppConfig

type AppConfig struct {
	Port    string
	GinMode string
	APIToken string

	// 腾讯云
	TencentEnabled      bool
	TencentCloudSecretId  string
	TencentCloudSecretKey string

	// 火山引擎
	VolcanoEnabled   bool
	VolcanoAccessKey string
	VolcanoSecretKey string
}

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("[config] .env 文件不存在，将使用系统环境变量")
	}

	// 腾讯云
	tencentEnabled := getEnv("TENCENTCLOUD_ENABLED", "false")
	tencentId := getEnv("TENCENTCLOUD_SECRET_ID", "")
	tencentKey := getEnv("TENCENTCLOUD_SECRET_KEY", "")
	volcanoEnabled := getEnv("VOLCANO_ENABLED", "false")
	volcanoKey := getEnv("VOLCANO_SECRET_KEY", "")
	volcanoAk := getEnv("VOLCANO_ACCESS_KEY", "")

	Config = AppConfig{
		Port:    getEnv("PORT", "8000"),
		GinMode: getEnv("GIN_MODE", "debug"),
		APIToken: getEnv("API_TOKEN", ""),

		TencentEnabled:      tencentEnabled == "true" && tencentId != "" && tencentKey != "",
		TencentCloudSecretId:  tencentId,
		TencentCloudSecretKey: tencentKey,

		VolcanoEnabled:   volcanoEnabled == "true" && volcanoKey != "" && volcanoAk != "",
		VolcanoAccessKey: volcanoAk,
		VolcanoSecretKey: volcanoKey,
	}

	if Config.APIToken == "" {
		log.Fatal("[config] API_TOKEN 未配置，请设置环境变量")
	}

	// 至少要启用一个 provider
	hasProvider := Config.TencentEnabled || Config.VolcanoEnabled
	if !hasProvider {
		log.Fatal("[config] 至少需要启用一个翻译服务（腾讯云或火山引擎）")
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}