package tencent

import (
	"context"
	"errors"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	mps "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mps/v20190612"
)

const (
	MaxTextLength = 2000 // 单次请求最大字符数（Unicode码点）
)

type Client struct {
	client *mps.Client
}

func NewClient(secretId, secretKey, region string) (*Client, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqTimeout = 30

	credential := common.NewCredential(secretId, secretKey)
	client, err := mps.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

// TranslateRequest 翻译请求
type TranslateRequest struct {
	Text       string
	SourceLang string // 源语言，支持 "auto" 自动识别
	TargetLang string // 目标语言
}

// TranslateResult 翻译结果
type TranslateResult struct {
	Text       string // 翻译后文本
	SourceLang string // 实际源语言（腾讯云返回）
	TargetLang string // 目标语言
	RequestId  string // 请求ID，用于日志追踪
}

// Translate 执行翻译
func (c *Client) Translate(ctx context.Context, req TranslateRequest) (*TranslateResult, error) {
	if len(req.Text) > MaxTextLength {
		return nil, errors.New("text exceeds maximum length of 2000 characters")
	}

	mpsReq := mps.NewTextTranslationRequest()
	mpsReq.SourceText = &req.Text
	mpsReq.Source = &req.SourceLang
	mpsReq.Target = &req.TargetLang

	resp, err := c.client.TextTranslationWithContext(ctx, mpsReq)
	if err != nil {
		return nil, parseError(err)
	}

	if resp.Response == nil {
		return nil, errors.New("empty response from Tencent Cloud")
	}

	result := &TranslateResult{
		Text:       *resp.Response.TargetText,
		SourceLang: *resp.Response.Source,
		TargetLang: *resp.Response.Target,
	}
	if resp.Response.RequestId != nil {
		result.RequestId = *resp.Response.RequestId
	}
	return result, nil
}

// parseError 将腾讯云 SDK 错误解析为用户友好的错误消息
func parseError(err error) error {
	if sdkErr, ok := err.(*tcerrors.TencentCloudSDKError); ok {
		// 保留腾讯云原始错误码和消息
		return errors.New(sdkErr.Code + ": " + sdkErr.Message)
	}
	return err
}

// IsTextTooLongError 判断是否是文本过长错误
func IsTextTooLongError(err error) bool {
	return strings.Contains(err.Error(), "TextTooLong") ||
		strings.Contains(err.Error(), "UnsupportedOperation")
}