package translate

import (
	"context"
	"errors"
	"testing"

	"FreeTranslate/internal/provider"
)

type dispatcherTestProvider struct {
	name string
	err  error
}

func (p *dispatcherTestProvider) Name() string { return p.name }
func (p *dispatcherTestProvider) Translate(context.Context, provider.Request) (*provider.Result, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &provider.Result{Text: p.name, SourceLang: "en", TargetLang: "zh-CN"}, nil
}
func (p *dispatcherTestProvider) TranslateBatch(ctx context.Context, reqs []provider.Request) ([]*provider.Result, []error) {
	results := make([]*provider.Result, len(reqs))
	errs := make([]error, len(reqs))
	for i, req := range reqs {
		results[i], errs[i] = p.Translate(ctx, req)
	}
	return results, errs
}
func (p *dispatcherTestProvider) MaxTextLen() int               { return 1000 }
func (p *dispatcherTestProvider) IsTextTooLongError(error) bool { return false }

func TestTranslateWithFailover(t *testing.T) {
	provider.Clear()
	defer provider.Clear()
	provider.Register(&dispatcherTestProvider{
		name: "first",
		err:  provider.NewProviderErrorWithPolicy("first", provider.ErrorUnavailable, "", "first unavailable", "", "", true, true, errors.New("down")),
	})
	provider.Register(&dispatcherTestProvider{name: "second"})

	result, used, err := translateWithFailover(context.Background(), provider.Request{
		Text: "hello", SourceLang: "en", TargetLang: "zh-CN",
	}, "auto")
	if err != nil {
		t.Fatal(err)
	}
	if used != "second" || result.Text != "second" {
		t.Fatalf("used provider = %q, result = %#v; want second", used, result)
	}
}
