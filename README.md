# FreeTranslate

基于云厂商免费额度的翻译 API 聚合服务。支持腾讯云、火山引擎、阿里云，按需启用。FreeTranslate 不提供公共翻译额度，实际可用额度和费用取决于调用方配置的云服务账号。

## 功能特性

- **多 Provider 支持**：腾讯云 TMT + 火山引擎 + 阿里云机器翻译，按需启用
- **灵活路由**：通过 `provider` 参数指定或自动选择
- **统一接口**：对外一个 API，内部自动适配各 Provider 的请求/响应格式
- **Provider 插拔架构**：新增 Provider 只需实现接口 + 注册，无需改动核心代码
- **Bearer Token 鉴权**
- **Zap 日志**（记录 Provider、语言、字符数和请求结果）

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置

```bash
cp .env.example .env
```

编辑 `.env` 填入各 Provider 的密钥，至少配置一个。

### 3. 启动

```bash
go run .
# 或编译后运行
go build -o freetranslate . && ./freetranslate
```

服务默认监听 `:8000`，可通过 `PORT` 环境变量修改。

### 4. 运行集成测试

确保服务已经启动后，在另一个终端执行：

```bash
go run cmd/test.go
```

测试工具会自动读取当前目录 `.env` 中的 `API_TOKEN` 和 Provider 开关，只测试已启用的 Provider。也可以通过环境变量覆盖：

```bash
BASE_URL=http://127.0.0.1:8000 \
TOKEN=your-api-token \
TEST_PROVIDERS=tencent,volcano \
go run cmd/test.go
```

测试会校验 HTTP 状态码、业务错误码、返回文本、语言代码、Provider、批量结果数量和索引。它不评价不同服务商之间的译文风格差异，也不会主动制造真实的云服务超时或额度耗尽。

## API 文档

### 健康检查

```
GET /health
```

无需鉴权，返回 `{"status": "ok"}`。

### 文本翻译

```
POST /v1/translate
Authorization: Bearer <your-api-token>
Content-Type: application/json
```

**请求体：**

| 字段 | 类型   | 必填 | 说明 |
|------|--------|------|------|
| `text` | string | ✅    | 待翻译文本 |
| `source_lang` | string | ❌    | 源语言代码，支持 `zh-CN`、`zh-TW`、`en` 等统一代码；不填或传 `auto` 时自动检测 |
| `target_lang` | string | ✅    | 目标语言代码，支持统一语言代码 |
| `provider` | string | ❌    | 指定 Provider，可选：`auto`（默认）、`tencent`、`volcano`、`alibaba-general` |

**响应示例：**

```json
{
  "code": 10000,
  "msg": "success",
  "data": {
    "text": "你好，世界！",
    "source_lang": "en",
    "target_lang": "zh-CN",
    "provider": "tencent"
  }
}
```

**调用示例：**

```bash
curl -X POST http://127.0.0.1:8000/v1/translate \
  -H 'Authorization: Bearer your-api-token' \
  -H 'Content-Type: application/json' \
  -d '{"text":"Hello, world!","source_lang":"en","target_lang":"zh-CN","provider":"auto"}'
```

> FreeTranslate 对外使用统一语言代码，内部会根据实际 Provider 自动转换。支持的公共语言包括：`zh-CN`、`zh-TW`、`en`、`ja`、`ko`、`fr`、`es`、`it`、`de`、`tr`、`ru`、`pt`、`vi`、`id`、`th`、`ms`、`ar`、`hi`。不同 Provider 对具体语言对的支持可能不同；`auto` 模式会在可切换错误时尝试其他已启用 Provider。

**语言支持：**
- 腾讯云：https://cloud.tencent.com/document/product/862/126431
- 火山引擎：https://docs.volcengine.com/docs/4640/65067
- 阿里云：https://help.aliyun.com/zh/machine-translation/developer-reference/machine-translation-language-code-list

**错误响应示例：**

```json
// 参数缺失
{"code": 40001, "msg": "target_lang is required"}

// 未知 Provider
{"code": 40010, "msg": "unknown provider: unknown, available: tencent, volcano, alibaba-general"}

// 不支持的统一语言代码
{"code": 40011, "type": "UNSUPPORTED_LANGUAGE", "msg": "unsupported language: xx-xx"}

// 文本超长（腾讯云 2000 / 火山引擎、阿里云通用版 5000）
{"code": 42200, "type": "TEXT_TOO_LONG", "msg": "文本长度超过服务商限制（最多 2000 个字符）", "provider": "tencent"}

// Provider 超时，可切换或重试
{"code": 50400, "type": "PROVIDER_TIMEOUT", "msg": "服务商响应超时", "provider": "tencent", "retryable": true, "request_id": "xxx"}

// 所有 Provider 都失败
{"code": 50300, "type": "ALL_PROVIDERS_FAILED", "msg": "所有翻译服务均不可用，请稍后重试", "retryable": true}
```

错误类型包括：`INVALID_ARGUMENT`、`UNSUPPORTED_LANGUAGE`、`TEXT_TOO_LONG`、`PROVIDER_TIMEOUT`、`PROVIDER_RATE_LIMITED`、`PROVIDER_UNAVAILABLE`、`PROVIDER_UNAUTHORIZED`、`TRANSLATION_FAILED` 和 `ALL_PROVIDERS_FAILED`。`retryable` 表示调用方稍后重试是否有意义；`auto` 是否切换 Provider 由服务内部根据 Provider 错误策略决定。

### 批量翻译

```
POST /v1/translate/batch
Authorization: Bearer <your-api-token>
Content-Type: application/json
```

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `texts` | []object | ✅ | 待翻译项列表，最多 100 条，每条独立指定目标语言 |
| `provider` | string | ❌ | `auto`（默认）、`tencent`、`volcano`、`alibaba-general` |

`texts` 每项结构：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `text` | string | ✅ | 待翻译文本 |
| `source_lang` | string | ❌    | 源语言代码，支持统一语言代码；不填或传 `auto` 时自动检测 |
| `target_lang` | string | ✅    | 目标语言代码，支持统一语言代码 |

**响应示例：**

```json
{
  "code": 10000,
  "msg": "success",
  "data": {
    "results": [
      {"index": 0, "text": "你好", "source_lang": "en", "target_lang": "zh-CN", "provider": "tencent"},
      {"index": 1, "text": "世界", "source_lang": "en", "target_lang": "zh-CN", "provider": "tencent"}
    ]
  }
}
```

错误时对应条目会填充统一的 `error`、`error_type`、`retryable`、`request_id` 和 `provider` 字段，其他条目正常返回。批量接口整体仍返回 HTTP 200，调用方需要逐条检查 `error` 字段。

```json
{
  "code": 10000,
  "msg": "success",
  "data": {
    "results": [
      {"index": 0, "text": "你好", "target_lang": "zh-CN", "provider": "tencent"},
      {"index": 1, "error": "服务商请求频率受限", "error_type": "PROVIDER_RATE_LIMITED", "retryable": true, "provider": "tencent"}
    ]
  }
}
```

`auto` 模式下，每条失败项会独立尝试后续 Provider，不会因为一条失败而阻塞其他条目。

---

## 配置说明

### Provider 开关

Provider 只有在开关为 `true` 且对应密钥完整时才会启用。仅配置密钥但开关为 `false`，或开关为 `true` 但密钥缺失，都会视为未启用。配置示例见 `.env.example`：

| Provider | 开关 | 说明 |
|----------|------|------|
| 腾讯云 | `TENCENTCLOUD_ENABLED=true` | `true`=启用（需同时配置密钥），`false`=禁用 |
| 火山引擎 | `VOLCANO_ENABLED=false` | `true`=启用（需同时配置 AK/SK），`false`=禁用 |
| 阿里云 | `ALIBABA_ENABLED=false` | `true`=启用（需同时配置 AK/SK），`false`=禁用 |

`API_TOKEN` 是 FreeTranslate 自身的访问 Token，不是云厂商密钥。请替换 `.env.example` 中的占位值，并不要提交 `.env`。

`provider` 参数指定具体 Provider。传 `auto` 或不传时，服务会从当前轮询位置开始，按注册顺序尝试已启用 Provider。正常请求会使用第一个成功的 Provider；遇到超时、限流、额度耗尽、服务不可用或语言不支持等可切换错误时，会自动尝试下一个 Provider。指定具体 Provider 时不会切换。所有 Provider 都失败时返回 `ALL_PROVIDERS_FAILED`。

---

## 腾讯云 TMT 文本翻译

### 开通服务

1. 登录 [腾讯云控制台 → 机器翻译](https://console.cloud.tencent.com/tmt)
2. 点击 **立即开通**
3. 获取 SecretId + SecretKey：[访问密钥管理](https://console.cloud.tencent.com/cam/capi)

### 子账户权限配置（CAM）

若使用子账户密钥，需授予 TMT 权限：

[CAM → 用户](https://console.cloud.tencent.com/cam/user) → 选择子用户 → **添加权限** → 搜索 `QcloudTMTFullAccess` → 关联

详细说明：[TMT 授权文档](https://cloud.tencent.com/document/product/551/52612)

### 计费

文档：[机器翻译计费概述](https://cloud.tencent.com/document/product/551/35017)

---

## 火山引擎机器翻译

### 开通服务

1. 登录 [火山引擎控制台 → 机器翻译](https://console.volcengine.com/translate)
2. 点击 **立即开通**
3. 获取 AccessKey + SecretKey：[密钥管理](https://console.volcengine.com/iam/keymanage)

### 子账户权限配置（IAM）

1. [IAM → 身份管理 → 用户](https://console.volcengine.com/iam/identitymanage/user) → 选择子用户
2. **权限** → **添加权限** → 搜索 `TranslateFullAccess` → 勾选 → 确认

### 计费

文档：[机器翻译产品计费](https://docs.volcengine.com/docs/4640/68515)

---

## 阿里云机器翻译

### 开通服务

1. 登录 [阿里云控制台 → 机器翻译](https://www.aliyun.com/product/ai/alimt)
2. 点击 **立即开通**
3. 获取 AccessKey + SecretKey：[AccessKey 管理](https://ram.console.aliyun.com/profile/access-keys)

### 子账户权限配置（RAM）

[RAM 控制台 → 授权](https://ram.console.aliyun.com/permissions) → **新增授权** → 搜索 `AliyunMTFullAccess` → 勾选 → 确认

### 计费

文档：[机器翻译计费概述](https://help.aliyun.com/zh/machine-translation/product-overview/billing-overview)

### 关于通用版与专业版

阿里云机器翻译分为**通用版**和**场景版**（专业版），两者是独立的 API：

| 版本 | Action | 场景 | 说明 |
|---|---|---|---|
| 通用版 | `TranslateGeneral` | `general` | 通用文本翻译 |
| 场景版 | `Translate` | `title`/`description`/`communication`/`medical`/`social`/`finance` | 垂直领域优化 |

FreeTranslate 当前接入的是**通用版**（`alibaba-general`）。如需场景版，可根据 [专业版调用文档](https://help.aliyun.com/zh/machine-translation/developer-reference/machine-translation-professional-call-guide) 自行扩展。

---

## 测试与验证

### 单元测试和静态检查

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

### 真实 API 集成测试

服务启动后执行：

```bash
go run cmd/test.go
```

集成测试会读取 `.env` 并根据 Provider 开关执行已启用服务商的单条、批量、语言归一化、`auto` 和基础错误场景。测试通过表示这些场景的接口链路和响应契约符合预期，不代表所有云厂商错误码、真实故障切换或译文质量都已覆盖。

## 项目结构

```
FreeTranslate/
├── main.go                              # 入口，初始化配置、日志、Provider 和 HTTP 服务
├── .env.example                         # 配置模板；实际 .env 不应提交
├── cmd/
│   └── test.go                          # 真实 API 集成测试工具
└── internal/
    ├── api/
    │   ├── middleware/auth.go           # Bearer Token 鉴权
    │   ├── routes/router.go             # 路由和健康检查
    │   └── translate/
    │       ├── handler.go               # 单条翻译接口和错误响应
    │       ├── batch.go                 # 批量翻译接口
    │       └── dispatcher.go            # Provider 轮询和故障切换
    ├── platform/
    │   ├── config/config.go             # .env 加载和 Provider 配置
    │   ├── gwe/response.go              # 统一响应结构
    │   └── logs/logger.go               # Zap 日志
    └── provider/
        ├── interface.go                 # Provider 接口定义
        ├── registry.go                  # 注册表和轮询顺序
        ├── errors.go                    # 公共错误模型和切换策略
        ├── language.go                  # 统一语言代码和 Provider 映射
        ├── tencent/
        │   ├── client.go                # 腾讯云实现
        │   └── errors.go                # 腾讯云错误码解析
        ├── volcano/
        │   ├── client.go                # 火山引擎实现
        │   └── errors.go                # 火山引擎错误码解析
        └── alibaba/
            ├── client.go                # 阿里云通用版实现
            └── errors.go                # 阿里云错误码解析
```

### 新增 Provider

1. 在 `internal/provider/<name>/` 下实现 `client.go` 和 Provider 专属错误解析。
2. 在 `main.go` 中初始化客户端并调用 `provider.Register()`。
3. 在 `.env.example` 和 `config.go` 中增加开关与密钥配置。
4. 在 `language.go` 中补充语言映射，并为错误码和核心行为增加测试。

---

## 相关链接

**腾讯云：**
- [TMT 官方文档](https://cloud.tencent.com/document/product/551) | [文本翻译 API](https://cloud.tencent.com/document/product/551/52610) | [机器翻译计费概述](https://cloud.tencent.com/document/product/551/35017) | [CAM 授权](https://cloud.tencent.com/document/product/551)

**火山引擎：**
- [机器翻译文档](https://docs.volcengine.com/docs/4640/62099) | [文本翻译 API](https://docs.volcengine.com/docs/4640/65067) | [产品计费](https://docs.volcengine.com/docs/4640/68515)

**阿里云：**
- [机器翻译文档](https://help.aliyun.com/zh/machine-translation/) | [通用版 API](https://help.aliyun.com/zh/machine-translation/developer-reference/api-reference-machine-translation-universal-version-call-guide) | [专业版 API](https://help.aliyun.com/zh/machine-translation/developer-reference/machine-translation-professional-call-guide) | [计费概述](https://help.aliyun.com/zh/machine-translation/product-overview/billing-overview)