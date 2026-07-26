package routes

import (
	"FreeTranslate/internal/api/middleware"
	"FreeTranslate/internal/api/translate"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Setup(handler *translate.Handler) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(requestLogger())

	// 健康检查（无需鉴权）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 翻译接口（需鉴权）
	v1 := r.Group("/v1")
	v1.Use(middleware.AuthToken())
	{
		v1.POST("/translate", handler.Translate)
		v1.POST("/translate/batch", handler.TranslateBatch)
	}

	return r
}

// requestLogger Gin 请求日志中间件
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始
		// 后续处理...

		c.Next()

		// 记录请求完成（由业务层记录实际日志）
		_ = zap.S()
	}
}