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
	if trans.DetectedSourceLanguage != nil {
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

// Ensure Client implements provider.Provider
var _ provider.Provider = (*Client)(nil)