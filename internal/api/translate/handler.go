package translate

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"FreeTranslate/internal/platform/gwe"
	"FreeTranslate/internal/platform/logs"
	"FreeTranslate/internal/provider"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TranslateRequest 翻译请求
type TranslateRequest struct {
	Text       string `json:"text" binding:"required"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang" binding:"required"`
	Provider   string `json:"provider"` // auto, tencent, volcano
}

// TranslateData 翻译响应数据
type TranslateData struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Provider   string `json:"provider"`
}

// Handler 翻译处理器
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// Translate 翻译接口
func (h *Handler) Translate(c *gin.Context) {
	var req TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40000, "invalid request body: "+err.Error())
		return
	}

	if req.TargetLang == "" {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40001, "target_lang is required")
		return
	}

	// 解析 provider
	pName := req.Provider
	if pName == "" {
		pName = "auto"
	}

	p, err := provider.GetOrDefault(pName)
	if err != nil {
		// 未知 provider
		if pName != "auto" {
			gwe.ErrorJSON(c, http.StatusBadRequest, 40010, "unknown provider: "+pName+", available: "+joinProviders())
			return
		}
		gwe.ErrorJSON(c, http.StatusServiceUnavailable, 50300, "no translation provider available")
		return
	}

	// 检查文本长度
	charCount := utf8.RuneCountInString(req.Text)
	if charCount > p.MaxTextLen() {
		gwe.ErrorJSON(c, http.StatusUnprocessableEntity, 42200,
			fmt.Sprintf("text exceeds maximum length of %d characters", p.MaxTextLen()))
		return
	}

	// 执行翻译
	result, err := p.Translate(c.Request.Context(), provider.Request{
		Text:       req.Text,
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
	})

	if err != nil {
		logs.Logger.Error("翻译失败",
			zap.String("provider", p.Name()),
			zap.String("source_lang", req.SourceLang),
			zap.String("target_lang", req.TargetLang),
			zap.Int("char_count", charCount),
			zap.String("error", err.Error()),
		)

		if p.IsTextTooLongError(err) {
			gwe.ErrorJSON(c, http.StatusUnprocessableEntity, 42200, "text exceeds maximum length")
			return
		}

		// 透传提供商错误
		gwe.ErrorJSON(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	// 成功日志
	logs.Logger.Info("翻译成功",
		zap.String("provider", p.Name()),
		zap.String("source_lang", result.SourceLang),
		zap.String("target_lang", result.TargetLang),
		zap.Int("char_count", charCount),
	)

	gwe.SuccessJSON(c, TranslateData{
		Text:       result.Text,
		SourceLang: result.SourceLang,
		TargetLang: result.TargetLang,
		Provider:   p.Name(),
	})
}

func joinProviders() string {
	names := provider.List()
	if len(names) == 0 {
		return "tencent, volcano"
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}