# FreeTranslate

免费翻译 API 聚合服务。支持腾讯云、火山引擎、阿里云多家服务商，按需启用。

## 功能特性

- **多 Provider 支持**：腾讯云 TMT + 火山引擎 + 阿里云机器翻译，按需启用
- **灵活路由**：通过 `provider` 参数指定或自动选择
- **统一接口**：对外一个 API，内部自动适配各 Provider 的请求/响应格式
- **Provider 插拔架构**：新增 Provider 只需实现接口 + 注册，无需改动核心代码
- **Bearer Token 鉴权**
- **Zap 日志**（含字符数、计费时长）

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
go run ./cmd/server
# 或编译后运行
go build -o freetranslate ./cmd/server && ./freetranslate
```

服务默认监听 `:8000`，可通过 `PORT` 环境变量修改。

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
| `source_lang` | string | ❌    | 源语言代码（腾讯云/阿里云支持 `auto`，火山引擎不支持） |
| `target_lang` | string | ✅    | 目标语言代码 |
| `provider` | string | ❌    | 指定 Provider，可选：`auto`（默认）、`tencent`、`volcano`、`alibaba-general` |

**响应示例：**

```json
{
  "code": 10000,
  "msg": "success",
  "data": {
    "text": "你好，世界！",
    "source_lang": "en",
    "target_lang": "zh",
    "provider": "tencent"
  }
}
```

> ⚠️ 各 Provider 的语言代码存在差异，使用前请参考语言支持列表。

**语言支持：**
- 腾讯云：https://cloud.tencent.com/document/product/862/126431
- 火山引擎：https://docs.volcengine.com/docs/4640/65067
- 阿里云：https://help.aliyun.com/zh/machine-translation/developer-reference/machine-translation-language-code-list

**错误响应示例：**

```json
// 参数缺失
{"code": 40001, "msg": "target_lang is required"}

// 未知 provider
{"code": 40010, "msg": "unknown provider: volcano, available: tencent, volcano, alibaba-general"}

// 文本超长（腾讯云/阿里云 2000 / 火山引擎 5000）
{"code": 42200, "msg": "text exceeds maximum length of 2000 characters"}

// Provider 错误透传
{"code": 50000, "msg": "UnsupportedOperation.TextTooLong: ..."}
```

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
| `source_lang` | string | ❌ | 源语言代码，不填则自动检测 |
| `target_lang` | string | ✅ | 目标语言代码 |

**响应示例：**

```json
{
  "code": 10000,
  "msg": "success",
  "data": {
    "results": [
      {"index": 0, "text": "你好", "source_lang": "en", "target_lang": "zh", "provider": "tencent"},
      {"index": 1, "text": "世界", "source_lang": "en", "target_lang": "zh", "provider": "tencent"}
    ]
  }
}
```

**错误时对应条目的 `error` 字段会填充错误信息，其他条目正常返回。**

---

## 配置说明

### Provider 开关

在 `.env` 中配置密钥即视为启用，也可以显式控制开关：

| Provider | 开关 | 说明 |
|----------|------|------|
| 腾讯云 | `TENCENTCLOUD_ENABLED=true` | `true`=启用（需同时配置密钥），`false`=禁用 |
| 火山引擎 | `VOLCANO_ENABLED=false` | `true`=启用（需同时配置 AK/SK），`false`=禁用 |
| 阿里云 | `ALIBABA_ENABLED=false` | `true`=启用（需同时配置 AK/SK），`false`=禁用 |

`provider` 参数指定具体 Provider，`auto` 或不传时按注册顺序选择。

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

## 项目结构

```
FreeTranslate/
├── cmd/server/main.go                   # 入口，初始化 Provider 并注册
├── .env / .env.example                  # 配置
└── internal/
    ├── api/
    │   ├── middleware/auth.go           # Bearer Token 鉴权
    │   ├── routes/router.go             # 路由
    │   └── translate/
    │       ├── handler.go               # 翻译接口（单条）
    │       └── batch.go                  # 批量翻译接口
    ├── platform/
    │   ├── config/config.go             # .env 加载 + Provider 配置
    │   ├── gwe/response.go              # 统一响应
    │   └── logs/logger.go               # Zap 日志
    └── provider/
            ├── interface.go                 # Provider 接口定义
            ├── registry.go                  # 注册表（全局 map）
            ├── tencent/
            │   └── client.go                # 腾讯云实现（SDK）
            ├── volcano/
            │   └── client.go                # 火山引擎实现（SDK）
            └── alibaba/
                        └── client.go                # 阿里云实现（SDK，通用版）
            ```

            **新增 Provider 步骤：**

            1. 在 `internal/provider/<name>/` 下实现 `client.go`，实现 `Provider` 接口
            2. 在 `main.go` 中 `NewClient` + `provider.Register()`
            3. 在 `.env` 添加开关和密钥
            4. 在 `config.go` 添加配置读取
```

---

## 相关链接

**腾讯云：**
- [TMT 官方文档](https://cloud.tencent.com/document/product/551) | [文本翻译 API](https://cloud.tencent.com/document/product/551/52610) | [机器翻译计费概述](https://cloud.tencent.com/document/product/551/35017) | [CAM 授权](https://cloud.tencent.com/document/product/551)

**火山引擎：**
- [机器翻译文档](https://docs.volcengine.com/docs/4640/62099) | [文本翻译 API](https://docs.volcengine.com/docs/4640/65067) | [产品计费](https://docs.volcengine.com/docs/4640/68515)

**阿里云：**
- [机器翻译文档](https://help.aliyun.com/zh/machine-translation/) | [通用版 API](https://help.aliyun.com/zh/machine-translation/developer-reference/api-reference-machine-translation-universal-version-call-guide) | [专业版 API](https://help.aliyun.com/zh/machine-translation/developer-reference/machine-translation-professional-call-guide) | [计费概述](https://help.aliyun.com/zh/machine-translation/product-overview/billing-overview)