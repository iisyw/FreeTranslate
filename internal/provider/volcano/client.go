package volcano

import (
	"context"
	"errors"

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
	textList := []*string{volcengine.String(req.Text)}
	input := &translate20250301.TranslateTextInput{
		TextList:       textList,
		TargetLanguage: volcengine.String(req.TargetLang),
	}
	if req.SourceLang != "" {
		input.SourceLanguage = volcengine.String(req.SourceLang)
	}

	output, err := c.client.TranslateTextWithContext(ctx, input)
	if err != nil {
		return nil, err
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
		result.SourceLang = *trans.DetectedSourceLanguage
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
	return false
}

// TranslateBatch 火山引擎支持真正的批量翻译，TextList 最多 16 条
// 每条错误单独返回，不阻塞其他
func (c *Client) TranslateBatch(ctx context.Context, reqs []provider.Request) ([]*provider.Result, []error) {
	// 火山引擎每次最多 16 条，超过则截断
	const maxBatch = 16

	results := make([]*provider.Result, len(reqs))
	errs := make([]error, len(reqs))

	for i, req := range reqs {
		if len(req.Text) > c.MaxTextLen() {
			errs[i] = errors.New("text exceeds maximum length")
			continue
		}
		textList := []*string{volcengine.String(req.Text)}
		input := &translate20250301.TranslateTextInput{
			TextList:       textList,
			TargetLanguage: volcengine.String(req.TargetLang),
		}
		if req.SourceLang != "" {
			input.SourceLanguage = volcengine.String(req.SourceLang)
		}

		output, err := c.client.TranslateTextWithContext(ctx, input)
		if err != nil {
			errs[i] = err
			continue
		}
		if output == nil || len(output.TranslationList) == 0 {
			errs[i] = errors.New("empty translation result")
			continue
		}

		trans := output.TranslationList[0]
		result := &provider.Result{
			Text:       volcengine.StringValue(trans.Translation),
			TargetLang: req.TargetLang,
		}
		if trans.DetectedSourceLanguage != nil && *trans.DetectedSourceLanguage != "" {
			result.SourceLang = *trans.DetectedSourceLanguage
		} else if req.SourceLang != "" {
			result.SourceLang = req.SourceLang
		} else {
			result.SourceLang = "auto"
		}
		if output.Metadata != nil && output.Metadata.RequestId != "" {
			result.RequestId = output.Metadata.RequestId
		}
		results[i] = result
	}

	return results, errs
}

// Ensure Client implements provider.Provider
var _ provider.Provider = (*Client)(nil)