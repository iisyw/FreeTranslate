package translate

import (
	"context"
	"testing"

	"FreeTranslate/internal/provider"
)

type batchDispatcherProvider struct {
	name string
	err  error
}

func (p *batchDispatcherProvider) Name() string { return p.name }
func (p *batchDispatcherProvider) Translate(context.Context, provider.Request) (*provider.Result, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &provider.Result{Text: p.name}, nil
}
func (p *batchDispatcherProvider) TranslateBatch(ctx context.Context, reqs []provider.Request) ([]*provider.Result, []error) {
	results := make([]*provider.Result, len(reqs))
	errs := make([]error, len(reqs))
	for i, req := range reqs {
		results[i], errs[i] = p.Translate(ctx, req)
	}
	return results, errs
}
func (p *batchDispatcherProvider) MaxTextLen() int               { return 1000 }
func (p *batchDispatcherProvider) IsTextTooLongError(error) bool { return false }

func TestTranslateBatchWithFailoverKeepsIndexes(t *testing.T) {
	provider.Clear()
	defer provider.Clear()
	provider.Register(&batchDispatcherProvider{
		name: "first",
		err:  provider.NewProviderErrorWithPolicy("first", provider.ErrorUnavailable, "", "unavailable", "", "", true, true, nil),
	})
	provider.Register(&batchDispatcherProvider{name: "second"})

	results, errs, providers := translateBatchWithFailover(context.Background(), []provider.Request{
		{Text: "one", SourceLang: "en", TargetLang: "zh-CN"},
		{Text: "two", SourceLang: "en", TargetLang: "zh-CN"},
	}, "auto")
	if errs[0] != nil || errs[1] != nil {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if providers[0] != "second" || providers[1] != "second" {
		t.Fatalf("providers = %v, want both second", providers)
	}
	if results[0].Text != "second" || results[1].Text != "second" {
		t.Fatalf("results = %#v", results)
	}
}
