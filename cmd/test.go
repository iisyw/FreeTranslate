// FreeTranslate API integration test tool.
// Start the service first, then run:
//   go run cmd/test.go
//
// Defaults to http://127.0.0.1:8000. Override with:
//   BASE_URL=http://localhost:8000 TOKEN=xxx go run cmd/test.go
//
// TOKEN takes precedence over API_TOKEN. When .env is present, API_TOKEN and
// provider enable flags are loaded automatically.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/godotenv/godotenv"
)

type reqOpts struct {
	url    string
	token  string
	client *http.Client
}

type translateReq struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang,omitempty"`
	TargetLang string `json:"target_lang"`
	Provider   string `json:"provider,omitempty"`
}

type batchReq struct {
	Texts    []translateReq `json:"texts"`
	Provider string         `json:"provider,omitempty"`
}

type apiResp struct {
	Code int             `json:"code"`
	Type string          `json:"type,omitempty"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data,omitempty"`
}

type translateData struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Provider   string `json:"provider"`
}

type batchResult struct {
	Index      int    `json:"index"`
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Provider   string `json:"provider"`
	Error      string `json:"error,omitempty"`
}

type callResult struct {
	status  int
	body    apiResp
	raw     []byte
	pretty  string
	err     error
	request string
}

var (
	passed int
	failed int
	red    = "\033[31m"
	green  = "\033[32m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

func main() {
	_ = godotenv.Load()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000"
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		token = os.Getenv("API_TOKEN")
	}
	if strings.TrimSpace(token) == "" {
		fmt.Fprintln(os.Stderr, "测试终止：未配置 TOKEN 或 API_TOKEN")
		os.Exit(2)
	}

	opts := &reqOpts{
		url:    strings.TrimRight(baseURL, "/"),
		token:  token,
		client: &http.Client{Timeout: 45 * time.Second},
	}
	providers := configuredProviders()
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "测试终止：没有检测到启用的 Provider，请设置 *_ENABLED=true 或 TEST_PROVIDERS")
		os.Exit(2)
	}

	section("健康检查")
	run("GET /health", true, func() (bool, string, string) {
		return health(opts)
	})

	for _, providerName := range providers {
		testProvider(opts, providerName)
	}

	section("自动选择 & 异常场景")
	run("auto 自动选择并校验返回结果", true, func() (bool, string, string) {
		return translate(opts, "auto", "Hello world", "", "zh", providers)
	})
	run("auto 批量翻译并校验每条结果", true, func() (bool, string, string) {
		return batchTranslate(opts, "auto", providers,
			translateReq{Text: "Hello", TargetLang: "zh"},
			translateReq{Text: "world", TargetLang: "zh"},
		)
	})
	run("未知 provider 返回 HTTP 400/code 40010", true, func() (bool, string, string) {
		return expectTranslateError(opts, "unknown", "Hello world", "", "zh", http.StatusBadRequest, 40010)
	})
	if contains(providers, "tencent") {
		run("腾讯云超长文本返回 HTTP 422/code 42200", true, func() (bool, string, string) {
			text := strings.Repeat("x", 2001)
			return expectTranslateError(opts, "tencent", text, "", "zh", http.StatusUnprocessableEntity, 42200)
		})
	} else {
		skip("腾讯云超长文本测试（腾讯云未启用）")
	}
	run("不支持语言返回 HTTP 400/code 40011", true, func() (bool, string, string) {
		return expectTranslateError(opts, "auto", "Hello", "", "xx-XX", http.StatusBadRequest, 40011)
	})

	fmt.Printf("\n%s====================%s\n", cyan, reset)
	fmt.Printf("%s  通过: %d  失败: %d%s\n", cyan, passed, failed, reset)
	fmt.Printf("%s====================%s\n", cyan, reset)
	if failed > 0 {
		os.Exit(1)
	}
}

func testProvider(opts *reqOpts, providerName string) {
	section(providerLabel(providerName))
	run("单条翻译(en→zh)并校验字段", true, func() (bool, string, string) {
		return translate(opts, providerName, "Good morning", "en", "zh", providersFor(providerName))
	})
	run("单条翻译(zh→en)并校验字段", true, func() (bool, string, string) {
		return translate(opts, providerName, "早上好", "zh", "en", providersFor(providerName))
	})
	run("zh-CN 归一化(target)", true, func() (bool, string, string) {
		return translate(opts, providerName, "Hello world", "", "zh-CN", providersFor(providerName))
	})
	run("en-US 归一化(source)", true, func() (bool, string, string) {
		return translate(opts, providerName, "Hello world", "en-US", "zh", providersFor(providerName))
	})
	run("批量翻译并校验条目数量、索引和文本", true, func() (bool, string, string) {
		return batchTranslate(opts, providerName, providersFor(providerName),
			translateReq{Text: "Hello", SourceLang: "en", TargetLang: "zh"},
			translateReq{Text: "world", SourceLang: "en", TargetLang: "zh"},
		)
	})
}

func configuredProviders() []string {
	if raw := strings.TrimSpace(os.Getenv("TEST_PROVIDERS")); raw != "" {
		return splitProviders(raw)
	}

	var providers []string
	if strings.EqualFold(os.Getenv("TENCENTCLOUD_ENABLED"), "true") {
		providers = append(providers, "tencent")
	}
	if strings.EqualFold(os.Getenv("VOLCANO_ENABLED"), "true") {
		providers = append(providers, "volcano")
	}
	if strings.EqualFold(os.Getenv("ALIBABA_ENABLED"), "true") {
		providers = append(providers, "alibaba-general")
	}
	return providers
}

func providersFor(providerName string) []string {
	return []string{providerName}
}

func splitProviders(raw string) []string {
	var providers []string
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimSpace(value)
		if value != "" && !contains(providers, value) {
			providers = append(providers, value)
		}
	}
	return providers
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func providerLabel(providerName string) string {
	switch providerName {
	case "tencent":
		return "腾讯云 TMT"
	case "volcano":
		return "火山引擎"
	case "alibaba-general":
		return "阿里云通用版"
	default:
		return providerName
	}
}

func section(title string) {
	fmt.Printf("\n%s=== %s ===%s\n", cyan, title, reset)
}

func run(desc string, _ bool, fn func() (bool, string, string)) {
	ok, req, resp := fn()
	if ok {
		fmt.Printf("  %s✓ PASS%s %s\n", green, reset, desc)
		passed++
	} else {
		fmt.Printf("  %s✗ FAIL%s %s\n", red, reset, desc)
		failed++
	}
	fmt.Printf("    %s请求:%s %s\n", yellow, reset, req)
	fmt.Printf("    %s响应:%s %s\n", yellow, reset, resp)
}

func skip(desc string) {
	fmt.Printf("  %s- SKIP%s %s\n", yellow, reset, desc)
}

func health(opts *reqOpts) (bool, string, string) {
	call := doRequest(opts, http.MethodGet, "/health", nil)
	if call.err != nil {
		return false, call.request, call.errorText()
	}
	var healthBody struct {
		Status string `json:"status"`
	}
	ok := call.status == http.StatusOK && json.Unmarshal(call.raw, &healthBody) == nil && healthBody.Status == "ok"
	return ok, call.request, call.responseText()
}

func translate(opts *reqOpts, providerName, text, sourceLang, targetLang string, acceptedProviders []string) (bool, string, string) {
	body := translateReq{
		Text:       text,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Provider:   providerName,
	}
	call := doJSON(opts, http.MethodPost, "/v1/translate", body)
	if call.err != nil {
		return false, call.request, call.errorText()
	}

	var data translateData
	ok := call.status == http.StatusOK && call.body.Code == 10000 && json.Unmarshal(call.body.Data, &data) == nil
	ok = ok && strings.TrimSpace(data.Text) != ""
	ok = ok && data.TargetLang == canonicalLanguage(targetLang)
	if sourceLang != "" && sourceLang != "auto" {
		ok = ok && data.SourceLang == canonicalLanguage(sourceLang)
	}
	if providerName == "auto" {
		ok = ok && contains(acceptedProviders, data.Provider)
	} else {
		ok = ok && data.Provider == providerName
	}
	return ok, call.request, call.responseText()
}

func batchTranslate(opts *reqOpts, providerName string, acceptedProviders []string, items ...translateReq) (bool, string, string) {
	call := doJSON(opts, http.MethodPost, "/v1/translate/batch", batchReq{Texts: items, Provider: providerName})
	if call.err != nil {
		return false, call.request, call.errorText()
	}

	var data struct {
		Results []batchResult `json:"results"`
	}
	ok := call.status == http.StatusOK && call.body.Code == 10000 && json.Unmarshal(call.body.Data, &data) == nil
	ok = ok && len(data.Results) == len(items)
	for i, result := range data.Results {
		if result.Index != i || strings.TrimSpace(result.Text) == "" || result.Error != "" {
			ok = false
		}
		if result.TargetLang != canonicalLanguage(items[i].TargetLang) {
			ok = false
		}
		if providerName == "auto" {
			ok = ok && contains(acceptedProviders, result.Provider)
		} else {
			ok = ok && result.Provider == providerName
		}
	}
	return ok, call.request, call.responseText()
}

func expectTranslateError(opts *reqOpts, providerName, text, sourceLang, targetLang string, expectedStatus, expectedCode int) (bool, string, string) {
	body := translateReq{
		Text:       text,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Provider:   providerName,
	}
	call := doJSON(opts, http.MethodPost, "/v1/translate", body)
	if call.err != nil {
		return false, call.request, call.errorText()
	}
	ok := call.status == expectedStatus && call.body.Code == expectedCode
	return ok, call.request, call.responseText()
}

func doJSON(opts *reqOpts, method, path string, body interface{}) callResult {
	payload, err := json.Marshal(body)
	if err != nil {
		return callResult{request: method + " " + path, err: err}
	}
	return doRequest(opts, method, path, payload)
}

func doRequest(opts *reqOpts, method, path string, payload []byte) callResult {
	call := callResult{request: method + " " + opts.url + path}
	request, err := http.NewRequest(method, opts.url+path, bytes.NewReader(payload))
	if err != nil {
		call.err = err
		return call
	}
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+opts.token)
	}

	response, err := opts.client.Do(request)
	if err != nil {
		call.err = err
		return call
	}
	defer response.Body.Close()
	call.status = response.StatusCode
	call.raw, call.err = io.ReadAll(response.Body)
	if call.err != nil {
		return call
	}
	call.pretty = string(call.raw)
	var pretty bytes.Buffer
	if json.Indent(&pretty, call.raw, "", "  ") == nil {
		call.pretty = pretty.String()
	}
	if err := json.Unmarshal(call.raw, &call.body); err != nil {
		call.err = err
	}
	return call
}

func (call callResult) responseText() string {
	return fmt.Sprintf("HTTP %d\n%s", call.status, call.pretty)
}

func (call callResult) errorText() string {
	if call.status != 0 {
		return fmt.Sprintf("HTTP %d\n%s\n请求错误: %v", call.status, call.pretty, call.err)
	}
	return fmt.Sprintf("请求错误: %v", call.err)
}

func canonicalLanguage(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "zh", "zh-cn", "zh-hans":
		return "zh-CN"
	case "zh-tw", "zh-hant", "zh-hant-tw":
		return "zh-TW"
	case "en-us", "en-gb":
		return "en"
	case "ja-jp":
		return "ja"
	case "ko-kr":
		return "ko"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}
