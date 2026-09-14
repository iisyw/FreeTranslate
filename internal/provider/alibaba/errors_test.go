package alibaba

import (
	"testing"

	"FreeTranslate/internal/provider"
)

func TestClassifyCode(t *testing.T) {
	tests := []struct {
		code string
		kind provider.ErrorKind
	}{
		{"10001", provider.ErrorTimeout},
		{"10005", provider.ErrorUnsupportedLanguage},
		{"10008", provider.ErrorTextTooLong},
		{"10009", provider.ErrorUnauthorized},
		{"10013", provider.ErrorUnavailable},
	}
	for _, tt := range tests {
		got := provider.AsProviderError(classifyCode(tt.code, "detail", "request-id", nil))
		if got == nil || got.Kind != tt.kind {
			if got == nil {
				t.Errorf("classifyCode(%q) returned nil", tt.code)
			} else {
				t.Errorf("classifyCode(%q).Kind = %q, want %q", tt.code, got.Kind, tt.kind)
			}
		}
	}
}
