package provider

import "testing"

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input  string
		source bool
		want   string
	}{
		{input: " zh-CN ", want: "zh-CN"},
		{input: "en-US", want: "en"},
		{input: "zh-hant-tw", want: "zh-TW"},
		{input: "auto", source: true, want: ""},
	}
	for _, tt := range tests {
		got, err := NormalizeLanguage(tt.input, tt.source)
		if err != nil {
			t.Fatalf("NormalizeLanguage(%q) returned error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("NormalizeLanguage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	if _, err := NormalizeLanguage("auto", false); err == nil {
		t.Fatal("target language auto should be rejected")
	}
	if _, err := NormalizeLanguage("zh-HK", false); err == nil {
		t.Fatal("unsupported regional language should be rejected")
	}
}

func TestProviderLanguageMapping(t *testing.T) {
	cases := map[string]string{
		"tencent":         "zh_TW",
		"volcano":         "zh-Hant-tw",
		"alibaba-general": "zh-TW",
	}
	for providerName, want := range cases {
		got, ok := ProviderLanguage(providerName, "zh-TW")
		if !ok || got != want {
			t.Errorf("ProviderLanguage(%q, zh-TW) = %q, %v; want %q, true", providerName, got, ok, want)
		}
	}
}
