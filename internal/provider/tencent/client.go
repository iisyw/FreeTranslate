package tencent

import (
	"context"
	"errors"
	"strings"

	"FreeTranslate/internal/provider"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	mps "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mps/v20190612"
)

const name = "tencent"

type Client struct {
	client *mps.Client
}

func NewClient(secretId, secretKey string) (*Client, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqTimeout = 30

	credential := common.NewCredential(secretId, secretKey)
	client, err := mps.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func (c *Client) Name() string    { return name }
func (c *Client) MaxTextLen() int { return 2000 }

func (c *Client) Translate(ctx context.Context, req provider.Request) (*provider.Result, error) {
	if len(req.Text) > c.MaxTextLen() {
		return nil, errors.New("text exceeds maximum length")
	}

	// source_lang 为空时默认 auto
	sourceLang := req.SourceLang
	if sourceLang == "" {
		sourceLang = "auto"
	}

	mpsReq := mps.NewTextTranslationRequest()
	mpsReq.SourceText = &req.Text
	mpsReq.Source = &sourceLang
	mpsReq.Target = &req.TargetLang

	resp, err := c.client.TextTranslationWithContext(ctx, mpsReq)
	if err != nil {
		return nil, parseError(err)
	}

	if resp.Response == nil {
		return nil, errors.New("empty response from Tencent Cloud")
	}

	result := &provider.Result{
		SourceLang: *resp.Response.Source,
		TargetLang: *resp.Response.Target,
	}
	if resp.Response.TargetText != nil {
		result.Text = *resp.Response.TargetText
	}
	if resp.Response.RequestId != nil {
		result.RequestId = *resp.Response.RequestId
	}
	return result, nil
}

func (c *Client) IsTextTooLongError(err error) bool {
	return strings.Contains(err.Error(), "TextTooLong") ||
		strings.Contains(err.Error(), "UnsupportedOperation")
}

func parseError(err error) error {
	if sdkErr, ok := err.(*tcerr.TencentCloudSDKError); ok {
		return errors.New(sdkErr.Code + ": " + sdkErr.Message)
	}
	return err
}

// Ensure Client implements provider.Provider
var _ provider.Provider = (*Client)(nil)