package translate

import (
	"net/http"
	"strings"
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

	if strings.TrimSpace(req.TargetLang) == "" {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40001, "target_lang is required")
		return
	}

	srcLang, err := provider.NormalizeLanguage(req.SourceLang, true)
	if err != nil {
		gwe.ErrorJSONWithType(c, http.StatusBadRequest, 40011, string(provider.ErrorUnsupportedLanguage), err.Error(), "", false, "")
		return
	}
	tgtLang, err := provider.NormalizeLanguage(req.TargetLang, false)
	if err != nil {
		gwe.ErrorJSONWithType(c, http.StatusBadRequest, 40011, string(provider.ErrorUnsupportedLanguage), err.Error(), "", false, "")
		return
	}

	pName := req.Provider
	if pName == "" {
		pName = "auto"
	}
	if pName != "auto" {
		if _, err := provider.Candidates(pName); err != nil {
			gwe.ErrorJSON(c, http.StatusBadRequest, 40010, "unknown provider: "+pName+", available: "+joinProviders())
			return
		}
	}
	result, usedProvider, err := translateWithFailover(c.Request.Context(), provider.Request{
		Text:       req.Text,
		SourceLang: srcLang,
		TargetLang: tgtLang,
	}, pName)
	charCount := utf8.RuneCountInString(req.Text)
	if err != nil {
		if logs.Logger != nil {
			logs.Logger.Error("翻译失败",
				zap.String("provider", usedProvider),
				zap.String("source_lang", srcLang),
				zap.String("target_lang", tgtLang),
				zap.Int("char_count", charCount),
				zap.String("error", formatProviderError(err)),
			)
		}
		writeProviderError(c, err, usedProvider)
		return
	}

	if logs.Logger != nil {
		logs.Logger.Info("翻译成功",
			zap.String("provider", usedProvider),
			zap.String("source_lang", result.SourceLang),
			zap.String("target_lang", result.TargetLang),
			zap.Int("char_count", charCount),
		)
	}

	gwe.SuccessJSON(c, TranslateData{
		Text:       result.Text,
		SourceLang: result.SourceLang,
		TargetLang: result.TargetLang,
		Provider:   usedProvider,
	})
}

func writeProviderError(c *gin.Context, err error, providerName string) {
	providerErr := provider.AsProviderError(err)
	if providerErr == nil {
		gwe.ErrorJSONWithType(c, http.StatusBadGateway, 50200, string(provider.ErrorTranslationFailed), "翻译服务处理失败", providerName, false, "")
		return
	}

	status := http.StatusBadGateway
	code := 50200
	switch providerErr.Kind {
	case provider.ErrorInvalidArgument, provider.ErrorUnsupportedLanguage:
		status = http.StatusBadRequest
		code = 40011
	case provider.ErrorTextTooLong:
		status = http.StatusUnprocessableEntity
		code = 42200
	case provider.ErrorTimeout:
		status = http.StatusGatewayTimeout
		code = 50400
	case provider.ErrorRateLimited:
		status = http.StatusTooManyRequests
		code = 42900
	case provider.ErrorAllProvidersFailed, provider.ErrorUnavailable:
		status = http.StatusServiceUnavailable
		code = 50300
	}
	if providerName == "" {
		providerName = providerErr.ProviderName
	}
	gwe.ErrorJSONWithType(c, status, code, string(providerErr.Kind), providerErr.Message, providerName, providerErr.Retryable, providerErr.RequestID)
}

func joinProviders() string {
	names := provider.List()
	if len(names) == 0 {
		return "tencent, volcano, alibaba-general"
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
