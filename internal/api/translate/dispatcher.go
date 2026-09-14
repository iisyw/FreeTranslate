package translate

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"FreeTranslate/internal/provider"
)

func translateWithFailover(ctx context.Context, req provider.Request, requestedProvider string) (*provider.Result, string, error) {
	candidates, err := provider.Candidates(requestedProvider)
	if err != nil {
		return nil, "", err
	}
	auto := requestedProvider == "" || requestedProvider == "auto"
	var lastErr error

	for _, p := range candidates {
		if utf8.RuneCountInString(req.Text) > p.MaxTextLen() {
			lastErr = provider.NewTextTooLongError(p.Name(), p.MaxTextLen())
			if !auto {
				return nil, p.Name(), lastErr
			}
			continue
		}

		result, callErr := p.Translate(ctx, req)
		if callErr == nil && result != nil {
			result.SourceLang = canonicalResultLanguage(p.Name(), result.SourceLang, req.SourceLang)
			result.TargetLang = req.TargetLang
			return result, p.Name(), nil
		}
		if callErr == nil {
			callErr = provider.NewProviderErrorWithPolicy(p.Name(), provider.ErrorUnavailable, "", "服务商未返回有效结果", "empty result", "", true, true, nil)
		}
		lastErr = callErr
		if !auto || !provider.CanFailover(callErr) {
			return nil, p.Name(), callErr
		}
	}

	if lastErr == nil {
		return nil, "", errors.New("no provider available")
	}
	if providerErr := provider.AsProviderError(lastErr); providerErr != nil &&
		(providerErr.Kind == provider.ErrorTextTooLong || providerErr.Kind == provider.ErrorUnsupportedLanguage) {
		return nil, providerErr.ProviderName, lastErr
	}
	return nil, "", provider.NewProviderErrorWithPolicy(
		"", provider.ErrorAllProvidersFailed, "",
		"所有翻译服务均不可用，请稍后重试", lastErr.Error(), "", false, true, lastErr,
	)
}

func translateBatchWithFailover(ctx context.Context, reqs []provider.Request, requestedProvider string) ([]*provider.Result, []error, []string) {
	results := make([]*provider.Result, len(reqs))
	errs := make([]error, len(reqs))
	providers := make([]string, len(reqs))
	pending := make([]int, len(reqs))
	for i := range reqs {
		pending[i] = i
	}

	candidates, candidateErr := provider.Candidates(requestedProvider)
	if candidateErr != nil {
		for i := range errs {
			errs[i] = candidateErr
		}
		return results, errs, providers
	}
	auto := requestedProvider == "" || requestedProvider == "auto"

	for _, p := range candidates {
		if len(pending) == 0 {
			break
		}
		localIndexes := make([]int, 0, len(pending))
		localReqs := make([]provider.Request, 0, len(pending))
		nextPending := make([]int, 0, len(pending))
		for _, index := range pending {
			if utf8.RuneCountInString(reqs[index].Text) > p.MaxTextLen() {
				errs[index] = provider.NewTextTooLongError(p.Name(), p.MaxTextLen())
				if auto {
					nextPending = append(nextPending, index)
				}
				continue
			}
			localIndexes = append(localIndexes, index)
			localReqs = append(localReqs, reqs[index])
		}

		if len(localReqs) > 0 {
			batchResults, batchErrs := p.TranslateBatch(ctx, localReqs)
			for localIndex, originalIndex := range localIndexes {
				var callErr error
				if localIndex < len(batchErrs) {
					callErr = batchErrs[localIndex]
				}
				if callErr == nil && localIndex < len(batchResults) && batchResults[localIndex] != nil {
					result := batchResults[localIndex]
					result.SourceLang = canonicalResultLanguage(p.Name(), result.SourceLang, reqs[originalIndex].SourceLang)
					result.TargetLang = reqs[originalIndex].TargetLang
					errs[originalIndex] = nil
					results[originalIndex] = result
					providers[originalIndex] = p.Name()
					continue
				}
				if callErr == nil {
					callErr = provider.NewProviderErrorWithPolicy(p.Name(), provider.ErrorUnavailable, "", "服务商未返回有效结果", "empty result", "", true, true, nil)
				}
				errs[originalIndex] = callErr
				if auto && provider.CanFailover(callErr) {
					nextPending = append(nextPending, originalIndex)
				}
			}
		}
		pending = nextPending
	}

	for _, index := range pending {
		if errs[index] == nil {
			errs[index] = provider.NewProviderErrorWithPolicy("", provider.ErrorAllProvidersFailed, "", "所有翻译服务均不可用，请稍后重试", "", "", false, true, nil)
		}
	}
	return results, errs, providers
}

func canonicalResultLanguage(providerName, actual, requested string) string {
	if actual == "" {
		if requested == "" {
			return "auto"
		}
		return requested
	}
	return provider.CanonicalLanguage(providerName, actual)
}

func formatProviderError(err error) string {
	if providerErr := provider.AsProviderError(err); providerErr != nil {
		return providerErr.Error()
	}
	return fmt.Sprintf("翻译服务处理失败: %v", err)
}
