package translate

import (
	"net/http"
	"unicode/utf8"

	"FreeTranslate/internal/platform/gwe"
	"FreeTranslate/internal/platform/logs"
	"FreeTranslate/internal/provider/tencent"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TranslateRequest 翻译请求
// @Summary 文本翻译
// @Description 将文本从源语言翻译到目标语言
// @Tags translate
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body TranslateRequest true "翻译请求"
// @Success 10000 {object} gwe.Response{data=TranslateData}
// @Failure 400 {object} gwe.ErrorResponse
// @Failure 401 {object} gwe.ErrorResponse
// @Failure 422 {object} gwe.ErrorResponse
// @Failure 500 {object} gwe.ErrorResponse
// @Router /v1/translate [post]
type TranslateRequest struct {
	Text       string `json:"text" binding:"required"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang" binding:"required"`
}

// TranslateData 翻译响应数据
type TranslateData struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

// Handler 处理翻译请求
type Handler struct {
	client *tencent.Client
}

func NewHandler(client *tencent.Client) *Handler {
	return &Handler{client: client}
}

// Translate 翻译接口
// @Summary 文本翻译
// @Description 调用腾讯云 MPS 文本翻译 API
func (h *Handler) Translate(c *gin.Context) {
	var req TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40000, "invalid request body: "+err.Error())
		return
	}

	// 默认 source_lang 为 auto
	if req.SourceLang == "" {
		req.SourceLang = "auto"
	}

	// 验证 target_lang
	if req.TargetLang == "" {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40001, "target_lang is required")
		return
	}

	// Unicode 码点字符数（腾讯云计费单位：1100字符/分钟）
	charCount := utf8.RuneCountInString(req.Text)

	result, err := h.client.Translate(c.Request.Context(), tencent.TranslateRequest{
		Text:       req.Text,
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
	})

	if err != nil {
		logs.Logger.Error("翻译失败",
			zap.String("source_lang", req.SourceLang),
			zap.String("target_lang", req.TargetLang),
			zap.Int("char_count", charCount),
			zap.String("error", err.Error()),
		)

		if tencent.IsTextTooLongError(err) {
			gwe.ErrorJSON(c, http.StatusUnprocessableEntity, 42200, "text exceeds maximum length of 2000 characters")
			return
		}
		// 透传腾讯云错误
		gwe.ErrorJSON(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	// 成功日志（带计费字段）
	logs.Logger.Info("翻译成功",
		zap.String("source_lang", result.SourceLang),
		zap.String("target_lang", result.TargetLang),
		zap.Int("char_count", charCount),
		zap.Float64("billable_minutes", float64(charCount)/1100.0),
	)

	gwe.SuccessJSON(c, TranslateData{
		Text:       result.Text,
		SourceLang: result.SourceLang,
		TargetLang: result.TargetLang,
	})
}