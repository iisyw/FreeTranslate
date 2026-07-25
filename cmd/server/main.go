package main

import (
	"net/http"
	"os"
	"time"

	"FreeTranslate/internal/api/routes"
	"FreeTranslate/internal/api/translate"
	"FreeTranslate/internal/platform/config"
	"FreeTranslate/internal/platform/logs"
	"FreeTranslate/internal/provider/tencent"

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

	// 3. 创建日志目录
	os.MkdirAll("logs", os.ModePerm)

	// 4. 初始化腾讯云翻译客户端
	client, err := tencent.NewClient(
		config.Config.TencentCloudSecretId,
		config.Config.TencentCloudSecretKey,
		config.Config.TencentCloudRegion,
	)
	if err != nil {
		logs.Logger.Fatal("初始化腾讯云翻译客户端失败", zap.Error(err))
	}

	// 5. 初始化 handler
	handler := translate.NewHandler(client)

	// 6. 设置 Gin 模式
	gin.SetMode(config.Config.GinMode)

	// 7. 设置路由
	router := routes.Setup(handler)

	// 8. 启动服务
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