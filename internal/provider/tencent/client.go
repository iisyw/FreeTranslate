package tencent

import (
	"context"
	"errors"
	"unicode/utf8"

	"FreeTranslate/internal/provider"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tmt "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tmt/v20180321"
)

const name = "tencent"

type Client struct {
	client *tmt.Client
}

func NewClient(secretId, secretKey string) (*Client, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqTimeout = 30

	credential := common.NewCredential(secretId, secretKey)
	client, err := tmt.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func (c *Client) Name() string    { return name }
func (c *Client) MaxTextLen() int { return 2000 }

func (c *Client) Translate(ctx context.Context, req provider.Request) (*provider.Result, error) {
	if utf8.RuneCountInString(req.Text) > c.MaxTextLen() {
		return nil, provider.NewTextTooLongError(name, c.MaxTextLen())
	}

	mapped, err := provider.MapRequest(name, req)
	if err != nil {
		return nil, classifyCode("UnsupportedOperation.UnsupportedLanguage", err.Error(), "", err)
	}

	sourceLang := mapped.SourceLang
	if sourceLang == "" {
		sourceLang = "auto"
	}

	tmtReq := tmt.NewTextTranslateRequest()
	tmtReq.SourceText = &req.Text
	tmtReq.Source = &sourceLang
	tmtReq.Target = &mapped.TargetLang
	tmtReq.ProjectId = common.Int64Ptr(0)

	resp, err := c.client.TextTranslateWithContext(ctx, tmtReq)
	if err != nil {
		return nil, classifyError(err)
	}

	if resp == nil || resp.Response == nil {
		return nil, classifyError(errors.New("empty response from Tencent Cloud TMT"))
	}

	result := &provider.Result{}
	if resp.Response.Source != nil {
		result.SourceLang = *resp.Response.Source
	}
	if resp.Response.Target != nil {
		result.TargetLang = *resp.Response.Target
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
	providerErr := provider.AsProviderError(err)
	return providerErr != nil && providerErr.Kind == provider.ErrorTextTooLong
}

// TranslateBatch 逐条调用 Translate
func (c *Client) TranslateBatch(ctx context.Context, reqs []provider.Request) ([]*provider.Result, []error) {
	results := make([]*provider.Result, len(reqs))
	errs := make([]error, len(reqs))
	for i, req := range reqs {
		results[i], errs[i] = c.Translate(ctx, req)
	}
	return results, errs
}

// Ensure Client implements provider.Provider
var _ provider.Provider = (*Client)(nil)
