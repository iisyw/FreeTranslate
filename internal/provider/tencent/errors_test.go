package tencent

import (
	"testing"

	"FreeTranslate/internal/provider"
)

func TestClassifyCode(t *testing.T) {
	tests := []struct {
		code  string
		kind  provider.ErrorKind
		fail  bool
		retry bool
	}{
		{"UnsupportedOperation.UnsupportedTargetLanguage", provider.ErrorUnsupportedLanguage, true, false},
		{"UnsupportedOperation.TextTooLong", provider.ErrorTextTooLong, true, false},
		{"FailedOperation.NoFreeAmount", provider.ErrorUnavailable, true, true},
		{"InternalError.BackendTimeout", provider.ErrorTimeout, true, true},
	}
	for _, tt := range tests {
		got := provider.AsProviderError(classifyCode(tt.code, "detail", "request-id", nil))
		if got == nil || got.Kind != tt.kind || got.CanFailover != tt.fail || got.Retryable != tt.retry {
			t.Errorf("classifyCode(%q) = %#v", tt.code, got)
		}
	}
}
