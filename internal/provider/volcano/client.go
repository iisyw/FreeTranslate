package volcano

import (
	"context"
	"errors"
	"unicode/utf8"

	"FreeTranslate/internal/provider"

	"github.com/volcengine/volcengine-go-sdk/service/translate20250301"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"
)

const name = "volcano"

type Client struct {
	client *translate20250301.TRANSLATE20250301
}

func NewClient(accessKey, secretKey string) *Client {
	cfg := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, "")).
		WithRegion("cn-north-1")

	sess, _ := session.NewSession(cfg)
	client := translate20250301.New(sess)
	return &Client{client: client}
}

func (c *Client) Name() string    { return name }
func (c *Client) MaxTextLen() int { return 5000 }

func (c *Client) Translate(ctx context.Context, req provider.Request) (*provider.Result, error) {
	if utf8.RuneCountInString(req.Text) > c.MaxTextLen() {
		return nil, provider.NewTextTooLongError(name, c.MaxTextLen())
	}

	mapped, err := provider.MapRequest(name, req)
	if err != nil {
		return nil, classifyCode("UnsupportedLanguage", err.Error(), "", err)
	}

	textList := []*string{volcengine.String(req.Text)}
	input := &translate20250301.TranslateTextInput{
		TextList:       textList,
		TargetLanguage: volcengine.String(mapped.TargetLang),
	}
	if mapped.SourceLang != "" {
		input.SourceLanguage = volcengine.String(mapped.SourceLang)
	}

	output, err := c.client.TranslateTextWithContext(ctx, input)
	if err != nil {
		return nil, classifyError(err)
	}

	if output == nil || len(output.TranslationList) == 0 {
		return nil, errors.New("empty translation result")
	}

	trans := output.TranslationList[0]
	result := &provider.Result{
		Text:       volcengine.StringValue(trans.Translation),
		TargetLang: req.TargetLang,
	}
	if trans.DetectedSourceLanguage != nil && *trans.DetectedSourceLanguage != "" {
		result.SourceLang = provider.CanonicalLanguage(name, *trans.DetectedSourceLanguage)
	} else if req.SourceLang != "" {
		result.SourceLang = req.SourceLang
	} else {
		result.SourceLang = "auto"
	}
	if output.Metadata != nil && output.Metadata.RequestId != "" {
		result.RequestId = output.Metadata.RequestId
	}

	return result, nil
}

func (c *Client) IsTextTooLongError(err error) bool {
	providerErr := provider.AsProviderError(err)
	return providerErr != nil && providerErr.Kind == provider.ErrorTextTooLong
}

// TranslateBatch translates each item independently so errors can be retried by the dispatcher.
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
