package provider

import (
	"fmt"
	"strings"
)

// Language describes one language exposed by the unified API.
type Language struct {
	Code    string
	Name    string
	Tencent string
	Volcano string
	Alibaba string
}

var languages = []Language{
	{Code: "zh-CN", Name: "简体中文", Tencent: "zh", Volcano: "zh", Alibaba: "zh"},
	{Code: "zh-TW", Name: "繁体中文", Tencent: "zh_TW", Volcano: "zh-Hant-tw", Alibaba: "zh-TW"},
	{Code: "en", Name: "英语", Tencent: "en", Volcano: "en", Alibaba: "en"},
	{Code: "ja", Name: "日语", Tencent: "ja", Volcano: "ja", Alibaba: "ja"},
	{Code: "ko", Name: "韩语", Tencent: "ko", Volcano: "ko", Alibaba: "ko"},
	{Code: "fr", Name: "法语", Tencent: "fr", Volcano: "fr", Alibaba: "fr"},
	{Code: "es", Name: "西班牙语", Tencent: "es", Volcano: "es", Alibaba: "es"},
	{Code: "it", Name: "意大利语", Tencent: "it", Volcano: "it", Alibaba: "it"},
	{Code: "de", Name: "德语", Tencent: "de", Volcano: "de", Alibaba: "de"},
	{Code: "tr", Name: "土耳其语", Tencent: "tr", Volcano: "tr", Alibaba: "tr"},
	{Code: "ru", Name: "俄语", Tencent: "ru", Volcano: "ru", Alibaba: "ru"},
	{Code: "pt", Name: "葡萄牙语", Tencent: "pt", Volcano: "pt", Alibaba: "pt"},
	{Code: "vi", Name: "越南语", Tencent: "vi", Volcano: "vi", Alibaba: "vi"},
	{Code: "id", Name: "印尼语", Tencent: "id", Volcano: "id", Alibaba: "id"},
	{Code: "th", Name: "泰语", Tencent: "th", Volcano: "th", Alibaba: "th"},
	{Code: "ms", Name: "马来语", Tencent: "ms", Volcano: "ms", Alibaba: "ms"},
	{Code: "ar", Name: "阿拉伯语", Tencent: "ar", Volcano: "ar", Alibaba: "ar"},
	{Code: "hi", Name: "印地语", Tencent: "hi", Volcano: "hi", Alibaba: "hi"},
}

var languageAliases = map[string]string{
	"zh":         "zh-CN",
	"zh-cn":      "zh-CN",
	"zh-hans":    "zh-CN",
	"zh-tw":      "zh-TW",
	"zh-hant":    "zh-TW",
	"zh-hant-tw": "zh-TW",
	"en-us":      "en",
	"en-gb":      "en",
	"ja-jp":      "ja",
	"ko-kr":      "ko",
	"fr-fr":      "fr",
	"de-de":      "de",
	"es-es":      "es",
	"pt-pt":      "pt",
	"it-it":      "it",
	"ru-ru":      "ru",
	"ar-ae":      "ar",
	"th-th":      "th",
	"vi-vn":      "vi",
	"id-id":      "id",
	"ms-my":      "ms",
	"tr-tr":      "tr",
}

// NormalizeLanguage validates and converts an external language code to the
// canonical code used by the API. An empty source language means auto-detect.
func NormalizeLanguage(value string, source bool) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if source && (value == "" || value == "auto") {
		return "", nil
	}
	if value == "" {
		return "", fmt.Errorf("language is required")
	}
	canonical, ok := languageAliases[value]
	if !ok {
		for _, language := range languages {
			if strings.EqualFold(language.Code, value) {
				canonical = language.Code
				ok = true
				break
			}
		}
	}
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", value)
	}
	return canonical, nil
}

// ProviderLanguage converts an API language code to the code expected by a
// specific provider.
func ProviderLanguage(providerName, canonical string) (string, bool) {
	for _, language := range languages {
		if language.Code != canonical {
			continue
		}
		switch providerName {
		case "tencent":
			return language.Tencent, language.Tencent != ""
		case "volcano":
			return language.Volcano, language.Volcano != ""
		case "alibaba-general":
			return language.Alibaba, language.Alibaba != ""
		default:
			return canonical, true
		}
	}
	return "", false
}

// MapRequest converts canonical API language codes for a Provider call.
func MapRequest(providerName string, req Request) (Request, error) {
	mapped := req
	if req.SourceLang != "" {
		code, ok := ProviderLanguage(providerName, req.SourceLang)
		if !ok {
			return Request{}, fmt.Errorf("unsupported source language: %s", req.SourceLang)
		}
		mapped.SourceLang = code
	}
	code, ok := ProviderLanguage(providerName, req.TargetLang)
	if !ok {
		return Request{}, fmt.Errorf("unsupported target language: %s", req.TargetLang)
	}
	mapped.TargetLang = code
	return mapped, nil
}

// CanonicalLanguage converts a provider response language code back to the API code.
func CanonicalLanguage(providerName, code string) string {
	if code == "" {
		return ""
	}
	for _, language := range languages {
		var providerCode string
		switch providerName {
		case "tencent":
			providerCode = language.Tencent
		case "volcano":
			providerCode = language.Volcano
		case "alibaba-general":
			providerCode = language.Alibaba
		}
		if strings.EqualFold(providerCode, code) {
			return language.Code
		}
	}
	return code
}
