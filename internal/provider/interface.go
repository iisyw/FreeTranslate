package provider

import (
	"context"
)

// Request 统一的翻译请求
type Request struct {
	Text       string
	SourceLang string // 空字符串表示 auto
	TargetLang string
}

// Result 统一的翻译结果
type Result struct {
	Text       string
	SourceLang string // 实际源语言（自动检测时返回检测结果）
	TargetLang string
	RequestId  string // 用于日志追踪
}

// Provider 翻译服务接口
type Provider interface {
	// Name 返回 provider 名称，如 "tencent" / "volcano"
	Name() string

	// Translate 执行翻译
	Translate(ctx context.Context, req Request) (*Result, error)

	// TranslateBatch 批量翻译（不支持则返回错误，由调用方降级为并发单次）
	// 每个元素的 error 也会单独返回，不阻塞其他元素
	TranslateBatch(ctx context.Context, reqs []Request) ([]*Result, []error)

	// MaxTextLen 最大文本长度限制（Unicode 字符）
	MaxTextLen() int

	// IsTextTooLongError 判断是否是文本过长错误
	IsTextTooLongError(err error) bool
}