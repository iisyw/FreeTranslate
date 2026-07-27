package translate

import (
	"net/http"

	"FreeTranslate/internal/platform/gwe"
	"FreeTranslate/internal/platform/logs"
	"FreeTranslate/internal/provider"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BatchTranslateRequest 批量翻译请求
type BatchTranslateRequest struct {
	Texts    []BatchTextItem `json:"texts" binding:"required"`
	Provider string          `json:"provider"`
}

// BatchTextItem 单条翻译项
type BatchTextItem struct {
	Text       string `json:"text" binding:"required"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang" binding:"required"`
}

// BatchTranslateData 批量翻译响应数据
type BatchTranslateData struct {
	Results []BatchResult `json:"results"`
}

// BatchResult 单条翻译结果
type BatchResult struct {
	Index     int    `json:"index"`
	Text      string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Provider  string `json:"provider"`
	Error     string `json:"error,omitempty"`
}

// TranslateBatch 批量翻译接口
func (h *Handler) TranslateBatch(c *gin.Context) {
	var req BatchTranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40000, "invalid request body: "+err.Error())
		return
	}

	if len(req.Texts) == 0 {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40002, "texts cannot be empty")
		return
	}

	if len(req.Texts) > 100 {
		gwe.ErrorJSON(c, http.StatusBadRequest, 40003, "texts cannot exceed 100 items")
		return
	}

	// 解析 provider
	pName := req.Provider
	if pName == "" {
		pName = "auto"
	}

	p, err := provider.GetOrDefault(pName)
	if err != nil {
		if pName != "auto" {
			gwe.ErrorJSON(c, http.StatusBadRequest, 40010, "unknown provider: "+pName+", available: "+joinProviders())
			return
		}
		gwe.ErrorJSON(c, http.StatusServiceUnavailable, 50300, "no translation provider available")
		return
	}

	// 构造批量请求（归一化语言码）
	reqs := make([]provider.Request, len(req.Texts))
	for i, item := range req.Texts {
		reqs[i] = provider.Request{
			Text:       item.Text,
			SourceLang: normalizeLang(item.SourceLang),
			TargetLang: normalizeLang(item.TargetLang),
		}
	}

	// 调用批量翻译
	results, errs := p.TranslateBatch(c.Request.Context(), reqs)

	// 整理结果
	output := make([]BatchResult, len(req.Texts))
	for i := range req.Texts {
		output[i] = BatchResult{
			Index: i,
		}
		if errs[i] != nil {
			output[i].Error = errs[i].Error()
			logs.Logger.Warn("批量翻译单条失败",
				zap.Int("index", i),
				zap.String("provider", p.Name()),
				zap.String("error", errs[i].Error()),
			)
		} else if results[i] != nil {
			output[i].Text = results[i].Text
			output[i].SourceLang = results[i].SourceLang
			output[i].TargetLang = results[i].TargetLang
			output[i].Provider = p.Name()
		}
	}

	logs.Logger.Info("批量翻译完成",
		zap.String("provider", p.Name()),
		zap.Int("total", len(req.Texts)),
		zap.Int("failed", countErrors(errs)),
	)

	gwe.SuccessJSON(c, BatchTranslateData{Results: output})
}

func countErrors(errs []error) int {
	n := 0
	for _, e := range errs {
		if e != nil {
			n++
		}
	}
	return n
}