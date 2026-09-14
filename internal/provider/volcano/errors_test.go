package volcano

import (
	"testing"

	"FreeTranslate/internal/provider"
)

func TestClassifyCode(t *testing.T) {
	tests := []struct {
		code string
		kind provider.ErrorKind
	}{
		{"FlowLimitExceeded", provider.ErrorRateLimited},
		{"InternalServiceTimeout", provider.ErrorTimeout},
		{"ServiceUnavailableTemp", provider.ErrorUnavailable},
		{"AccessDenied", provider.ErrorUnauthorized},
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
