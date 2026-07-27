package alibaba

import (
	"context"
	"errors"

	"FreeTranslate/internal/provider"

	alimt20181012 "github.com/alibabacloud-go/alimt-20181012/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
)

type Client struct {
	name      string
	client    *alimt20181012.Client
	scene     string
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
		name:      "alibaba-general",
		client:    aliClient,
		scene:     "general",
		maxTextLen: 2000,
	}, nil
}

func (c *Client) Name() string    { return c.name }
func (c *Client) MaxTextLen() int { return c.maxTextLen }

func (c *Client) Translate(ctx context.Context, req provider.Request) (*provider.Result, error) {
	input := &alimt20181012.TranslateRequest{
		SourceText:     tea.String(req.Text),
		TargetLanguage: tea.String(req.TargetLang),
		FormatType:     tea.String("text"),
		Scene:          tea.String(c.scene),
	}

	// source_lang 为空时传 auto
	srcLang := req.SourceLang
	if srcLang == "" {
		srcLang = "auto"
	}
	input.SourceLanguage = tea.String(srcLang)

	resp, err := c.client.Translate(input)
	if err != nil {
		return nil, err
	}

	result := &provider.Result{
		TargetLang: req.TargetLang,
		RequestId:  tea.StringValue(resp.Body.RequestId),
	}

	if resp.Body.Code != nil && *resp.Body.Code != 200 {
		return nil, errors.New(tea.StringValue(resp.Body.Message))
	}

	if resp.Body.Data != nil && resp.Body.Data.Translated != nil {
		result.Text = tea.StringValue(resp.Body.Data.Translated)
	}

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
		if len(req.Text) > c.maxTextLen {
			errs[i] = errors.New("text exceeds maximum length of 2000 characters")
			continue
		}
		srcLang := req.SourceLang
		if srcLang == "" {
			srcLang = "auto"
		}
		input := &alimt20181012.TranslateRequest{
			SourceText:     tea.String(req.Text),
			SourceLanguage: tea.String(srcLang),
			TargetLanguage: tea.String(req.TargetLang),
			FormatType:     tea.String("text"),
			Scene:          tea.String(c.scene),
		}
		resp, err := c.client.Translate(input)
		if err != nil {
			errs[i] = err
			continue
		}
		result := &provider.Result{
			TargetLang: req.TargetLang,
			RequestId:  tea.StringValue(resp.Body.RequestId),
		}
		if resp.Body.Code != nil && *resp.Body.Code != 200 {
			errs[i] = errors.New(tea.StringValue(resp.Body.Message))
			results[i] = result
			continue
		}
		if resp.Body.Data != nil && resp.Body.Data.Translated != nil {
			result.Text = tea.StringValue(resp.Body.Data.Translated)
		}
		if req.SourceLang != "" {
			result.SourceLang = req.SourceLang
		} else {
			result.SourceLang = "auto"
		}
		results[i] = result
	}

	return results, errs
}

func (c *Client) IsTextTooLongError(err error) bool {
	return false
}

// Ensure Client implements provider.Provider
var _ provider.Provider = (*Client)(nil)