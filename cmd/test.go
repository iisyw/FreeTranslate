// FreeTranslate 接口测试工具
// 用法：先启动服务，然后执行：
//   go run cmd/test.go
//
// 默认连接 http://127.0.0.1:8000，可通过环境变量覆盖：
//   BASE_URL=http://localhost:8000 TOKEN=xxx  go run cmd/test.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type reqOpts struct {
	url   string
	token string
}

type translateReq struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang,omitempty"`
	TargetLang string `json:"target_lang"`
	Provider   string `json:"provider"`
}

type batchReq struct {
	Texts    []translateReq `json:"texts"`
	Provider string         `json:"provider"`
}

type apiResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data,omitempty"`
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
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000"
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		token = "change-me-before-production"
	}
	opts := &reqOpts{url: baseURL, token: token}

	// 健康检查
	section("健康检查")
	run("GET /health", true, func() (bool, string, string) {
		return health(opts)
	})

	// 腾讯云
	section("腾讯云 TMT")
	run("单条翻译(en→zh)", true, func() (bool, string, string) {
		return translate(opts, "tencent", "Good morning", "", "zh")
	})
	run("单条翻译(zh→en)", true, func() (bool, string, string) {
		return translate(opts, "tencent", "早上好", "zh", "en")
	})
	run("zh-CN 归一化(target)", true, func() (bool, string, string) {
		return translate(opts, "tencent", "Hello world", "", "zh-CN")
	})
	run("en-US 归一化(source)", true, func() (bool, string, string) {
		return translate(opts, "tencent", "Hello world", "en-US", "zh")
	})
	run("批量翻译(2条)", true, func() (bool, string, string) {
		return batchTranslate(opts, "tencent",
			translateReq{Text: "Hello", TargetLang: "zh"},
			translateReq{Text: "world", TargetLang: "zh"},
		)
	})

	// 火山引擎
	section("火山引擎")
	run("单条翻译(en→zh)", true, func() (bool, string, string) {
		return translate(opts, "volcano", "Hello world", "", "zh")
	})
	run("批量翻译(2条)", true, func() (bool, string, string) {
		return batchTranslate(opts, "volcano",
			translateReq{Text: "Hello", SourceLang: "en", TargetLang: "zh"},
			translateReq{Text: "world", SourceLang: "en", TargetLang: "zh"},
		)
	})

	// 阿里云
	section("阿里云")
	run("单条翻译(en→zh)", true, func() (bool, string, string) {
		return translate(opts, "alibaba-general", "Hello world", "", "zh")
	})
	run("批量翻译(2条)", true, func() (bool, string, string) {
		return batchTranslate(opts, "alibaba-general",
			translateReq{Text: "Hello", SourceLang: "en", TargetLang: "zh"},
			translateReq{Text: "world", SourceLang: "en", TargetLang: "zh"},
		)
	})

	// auto + 异常
	section("自动选择 & 异常场景")
	run("auto 自动选择", true, func() (bool, string, string) {
		return translate(opts, "auto", "Hello world", "", "zh")
	})
	run("auto 批量翻译", true, func() (bool, string, string) {
		return batchTranslate(opts, "auto",
			translateReq{Text: "Hello", TargetLang: "zh"},
			translateReq{Text: "world", TargetLang: "zh"},
		)
	})
	run("未知 provider 返回 40010", false, func() (bool, string, string) {
		return translate(opts, "unknown", "Hello world", "", "zh")
	})
	run("超长文本拒绝 42200", false, func() (bool, string, string) {
		text := strings.Repeat("x", 2001)
		return translate(opts, "tencent", text, "", "zh")
	})

	// 汇总
	fmt.Printf("\n%s====================%s\n", cyan, reset)
	fmt.Printf("%s  通过: %d  失败: %d%s\n", cyan, passed, failed, reset)
	fmt.Printf("%s====================%s\n", cyan, reset)
	if failed > 0 {
		os.Exit(1)
	}
}

func section(title string) {
	fmt.Printf("\n%s=== %s ===%s\n", cyan, title, reset)
}

func run(desc string, expectSuccess bool, fn func() (bool, string, string)) {
	ok, req, resp := fn()
	// 对于预期失败的场景，反向判断
	ok = ok == expectSuccess
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

func health(opts *reqOpts) (bool, string, string) {
	reqDesc := "GET " + opts.url + "/health"
	resp, err := http.Get(opts.url + "/health")
	if err != nil {
		return false, reqDesc, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode == 200, reqDesc, fmt.Sprintf("HTTP %d %s", resp.StatusCode, string(raw))
}

func translate(opts *reqOpts, provider, text, sourceLang, targetLang string) (bool, string, string) {
	body := translateReq{
		Text:       text,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Provider:   provider,
	}
	textDisp := text
	if len(textDisp) > 100 {
		textDisp = textDisp[:100] + fmt.Sprintf("...(%d chars)", len(text))
	}
	reqJSON, _ := json.Marshal(body)
	reqDesc := fmt.Sprintf(`POST /v1/translate {"text":%q,"source_lang":%q,"target_lang":%q,"provider":%q}`, textDisp, sourceLang, targetLang, provider)
	ok, _, respBody := doPost(opts, "/v1/translate", reqJSON)
	return ok, reqDesc, respBody
}

func batchTranslate(opts *reqOpts, provider string, items ...translateReq) (bool, string, string) {
	if items == nil {
		items = []translateReq{}
	}
	body := batchReq{Texts: items, Provider: provider}
	reqJSON, _ := json.Marshal(body)
	reqDesc := fmt.Sprintf("POST /v1/translate/batch %s", string(reqJSON))
	ok, _, respBody := doPost(opts, "/v1/translate/batch", reqJSON)
	return ok, reqDesc, respBody
}

func doPost(opts *reqOpts, path string, bodyJSON []byte) (bool, string, string) {
	req, _ := http.NewRequest("POST", opts.url+path, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+opts.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return false, fmt.Sprintf("HTTP %d %s", resp.StatusCode, string(raw)),
			fmt.Sprintf("HTTP %d (非JSON响应)", resp.StatusCode)
	}

	var r apiResp
	ok := false
	if err := json.Unmarshal(raw, &r); err == nil && r.Code == 10000 {
		ok = true
	}

	return ok, fmt.Sprintf("HTTP %d %s", resp.StatusCode, string(raw)),
		fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, pretty.String())
}
