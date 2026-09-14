package alibaba

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"FreeTranslate/internal/provider"

	alimt20181012 "github.com/alibabacloud-go/alimt-20181012/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
)

const name = "alibaba-general"

type Client struct {
	name       string
	client     *alimt20181012.Client
	scene      string
	maxTextLen int
}

func NewClient(accessKey, secretKey string) (*Client, error) {
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(secretKey),
		Endpoint:        tea.String("mt.aliyuncs.com"),
	}

	aliClient, err := alimt20181012.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		name:       name,
		client:     aliClient,
		scene:      "general",
		maxTextLen: 5000,
	}, nil
}

func (c *Client) Name() string    { return c.name }
func (c *Client) MaxTextLen() int { return c.maxTextLen }

func (c *Client) Translate(ctx context.Context, req provider.Request) (*provider.Result, error) {
	if utf8.RuneCountInString(req.Text) > c.maxTextLen {
		return nil, provider.NewTextTooLongError(c.name, c.maxTextLen)
	}

	mapped, err := provider.MapRequest(c.name, req)
	if err != nil {
		return nil, classifyCode("UnsupportedLanguage", err.Error(), "", err)
	}

	input := &alimt20181012.TranslateGeneralRequest{
		SourceText:     tea.String(req.Text),
		TargetLanguage: tea.String(mapped.TargetLang),
		FormatType:     tea.String("text"),
		Scene:          tea.String(c.scene),
	}

	// source_lang 为空时传 auto
	srcLang := mapped.SourceLang
	if srcLang == "" {
		srcLang = "auto"
	}
	input.SourceLanguage = tea.String(srcLang)

	resp, err := c.client.TranslateGeneral(input)
	if err != nil {
		return nil, classifyError(err)
	}
	if resp == nil || resp.Body == nil {
		return nil, classifyError(errors.New("empty response from Alibaba Cloud"))
	}

	result := &provider.Result{
		TargetLang: req.TargetLang,
		RequestId:  tea.StringValue(resp.Body.RequestId),
	}

	if resp.Body.Code != nil && *resp.Body.Code != 200 {
		code := fmt.Sprintf("%d", *resp.Body.Code)
		return nil, classifyCode(code, tea.StringValue(resp.Body.Message), tea.StringValue(resp.Body.RequestId), errors.New(tea.StringValue(resp.Body.Message)))
	}

	if resp.Body.Data == nil || resp.Body.Data.Translated == nil {
		return nil, classifyCode("", "empty translation result", tea.StringValue(resp.Body.RequestId), errors.New("empty translation result"))
	}
	result.Text = tea.StringValue(resp.Body.Data.Translated)

	if req.SourceLang != "" {
		result.SourceLang = req.SourceLang
	} else {
		result.SourceLang = "auto"
	}

	return result, nil
}

func (c *Client) TranslateBatch(ctx context.Context, reqs []provider.Request) ([]*provider.Result, []error) {
	results := make([]*provider.Result, len(reqs))
	errs := make([]error, len(reqs))
	for i, req := range reqs {
		results[i], errs[i] = c.Translate(ctx, req)
	}
	return results, errs
}

func (c *Client) IsTextTooLongError(err error) bool {
	providerErr := provider.AsProviderError(err)
	return providerErr != nil && providerErr.Kind == provider.ErrorTextTooLong
}

// Ensure Client implements provider.Provider
var _ provider.Provider = (*Client)(nil)
