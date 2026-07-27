package main

import (
	"net/http"
	"os"
	"time"

	"FreeTranslate/internal/api/routes"
	"FreeTranslate/internal/api/translate"
	"FreeTranslate/internal/platform/config"
	"FreeTranslate/internal/platform/logs"
	"FreeTranslate/internal/provider"
	"FreeTranslate/internal/provider/alibaba"
	"FreeTranslate/internal/provider/tencent"
	"FreeTranslate/internal/provider/volcano"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. 初始化配置
	config.Init()

	// 2. 初始化日志
	logs.Init(logs.LogConfig{
		Level:      "debug",
		MaxSize:    50,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
		DirPath:    "logs",
	})
	os.MkdirAll("logs", os.ModePerm)

	// 3. 初始化并注册 Provider
	if config.Config.TencentEnabled {
		client, err := tencent.NewClient(
			config.Config.TencentCloudSecretId,
			config.Config.TencentCloudSecretKey,
		)
		if err != nil {
			logs.Logger.Fatal("初始化腾讯云翻译客户端失败", zap.Error(err))
		}
		provider.Register(client)
		logs.Logger.Info("腾讯云翻译已启用", zap.String("name", client.Name()))
	}

	if config.Config.VolcanoEnabled {
		client := volcano.NewClient(
			config.Config.VolcanoAccessKey,
			config.Config.VolcanoSecretKey,
		)
		provider.Register(client)
		logs.Logger.Info("火山引擎翻译已启用", zap.String("name", client.Name()))
	}

	if config.Config.AlibabaEnabled {
		client, err := alibaba.NewClient(
			config.Config.AlibabaAccessKey,
			config.Config.AlibabaSecretKey,
		)
		if err != nil {
			logs.Logger.Fatal("初始化阿里云翻译客户端失败", zap.Error(err))
		}

		provider.Register(client)
		logs.Logger.Info("阿里云翻译已启用", zap.String("name", client.Name()))
	}

	registered := provider.List()
	logs.Logger.Info("已注册的 Provider", zap.Any("providers", registered))

	// 4. 初始化 Handler（无需传 client，通过注册表获取）
	handler := translate.NewHandler()

	// 5. 设置 Gin 模式
	gin.SetMode(config.Config.GinMode)

	// 6. 设置路由
	router := routes.Setup(handler)

	// 7. 启动服务
	addr := ":" + config.Config.Port
	logs.Logger.Info("服务启动", zap.String("addr", addr))

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logs.Logger.Fatal("服务启动失败", zap.Error(err))
	}

	defer logs.Logger.Sync()
}