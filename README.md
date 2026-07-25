# FreeTranslate

免费翻译 API 聚合服务。支持多家服务商，默认腾讯云，开箱即用。

## 功能特性

- **多 Provider 支持**：腾讯云 MPS + 火山引擎，按需启用
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
| `source_lang` | string | ❌    | 源语言代码（腾讯云支持 `auto`，火山引擎不支持） |
| `target_lang` | string | ✅    | 目标语言代码 |
| `provider` | string | ❌    | 指定 Provider，可选：`auto`（默认）、`tencent`、`volcano` |

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

> ⚠️ 腾讯云与火山引擎的语言代码存在差异，使用前请参考各 Provider 的语言支持列表。

**语言支持：**
- 腾讯云：https://cloud.tencent.com/document/product/862/126431
- 火山引擎：https://docs.volcengine.com/docs/4640/65067

**错误响应示例：**

```json
// 参数缺失
{"code": 40001, "msg": "target_lang is required"}

// 未知 provider
{"code": 40010, "msg": "unknown provider: volcano, available: tencent, volcano"}

// 文本超长（腾讯云 2000 / 火山引擎 5000）
{"code": 42200, "msg": "text exceeds maximum length of 2000 characters"}

// Provider 错误透传
{"code": 50000, "msg": "UnsupportedOperation.TextTooLong: ..."}
```

---

## 配置说明

### Provider 开关

在 `.env` 中配置密钥即视为启用，也可以显式控制开关：

| Provider | 开关 | 说明 |
|----------|------|------|
| 腾讯云 | `TENCENTCLOUD_ENABLED=true` | `true`=启用（需同时配置密钥），`false`=禁用 |
| 火山引擎 | `VOLCANO_ENABLED=false` | `true`=启用（需同时配置 AK/SK），`false`=禁用 |

`provider` 参数指定具体 Provider，`auto` 或不传时按注册顺序选择。

---

## 腾讯云 MPS 文本翻译

### 开通服务

1. 登录 [腾讯云控制台 → 机器翻译](https://console.cloud.tencent.com/tmt)
2. 点击 **立即开通**
3. 获取 SecretId + SecretKey：[访问密钥管理](https://console.cloud.tencent.com/cam/capi)

### 子账户权限配置（CAM）

若使用子账户密钥，需授予 MPS 权限：

#### 方式一：预设策略（开发测试用）

CAM → 用户 → 添加权限 → 搜索 `QcloudMPSFullAccess`

#### 方式二：最小权限策略（生产环境推荐）

1. [CAM → 策略](https://console.cloud.tencent.com/cam/policy) → **创建策略** → **使用策略语法创建**
2. 填入：

```json
{
  "version": "2.0",
  "statement": [{
    "effect": "allow",
    "action": ["mps:TextTranslation"],
    "resource": ["*"]
  }]
}
```

3. 关联到子账户

详细说明：[媒体处理账号授权文档](https://cloud.tencent.com/document/product/862/117336)

### 计费

文档：[机器翻译计费概述](https://cloud.tencent.com/document/product/551/35017)

---

## 火山引擎机器翻译

### 开通服务

1. 登录 [火山引擎控制台 → 机器翻译](https://console.volcengine.com/translate)
2. 点击 **立即开通**
3. 获取 AccessKey + SecretKey：[密钥管理](https://console.volcengine.com/iam/keymanage)

### 子账户权限配置（IAM）

#### 方式一：系统预设策略（推荐开发测试用）

1. [IAM → 身份管理 → 用户](https://console.volcengine.com/iam/identitymanage/user) → 选择子用户
2. **权限** → **添加权限** → 搜索 `TranslateFullAccess` → 勾选 → 确认

#### 方式二：最小权限策略（生产环境推荐）

1. [IAM → 权限策略](https://console.volcengine.com/iam/policy) → **新建策略**
2. 选择 **策略语法** → **空白模板**
3. 填入以下内容，策略名称如 `Translate-TextTranslation`：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "translate:TranslateText"
      ],
      "Resource": ["*"]
    }
  ]
}
```

4. 保存后关联到子用户

### 计费

文档：[机器翻译产品计费](https://docs.volcengine.com/docs/4640/68515)

---

## 项目结构

```
FreeTranslate/
├── cmd/server/main.go                   # 入口，初始化 Provider 并注册
├── .env / .env.example                  # 配置
└── internal/
    ├── api/
    │   ├── middleware/auth.go           # Bearer Token 鉴权
    │   ├── routes/router.go            # 路由
    │   └── translate/handler.go         # 翻译接口（统一入口）
    ├── platform/
    │   ├── config/config.go             # .env 加载 + Provider 配置
    │   ├── gwe/response.go             # 统一响应
    │   └── logs/logger.go              # Zap 日志
    └── provider/
        ├── interface.go                 # Provider 接口定义
        ├── registry.go                  # 注册表（全局 map）
        ├── tencent/
        │   └── client.go                # 腾讯云实现（SDK）
        └── volcano/
            └── client.go                # 火山引擎实现（SDK）
```

**新增 Provider 步骤：**

1. 在 `internal/provider/<name>/` 下实现 `client.go`，实现 `Provider` 接口
2. 在 `main.go` 中 `NewClient` + `provider.Register()`
3. 在 `.env` 添加开关和密钥
4. 在 `config.go` 添加配置读取

---

## 相关链接

**腾讯云：**
- [MPS 官方文档](https://cloud.tencent.com/document/product/862) | [文本翻译 API](https://cloud.tencent.com/document/product/862/126431) | [机器翻译计费概述](https://cloud.tencent.com/document/product/551/35017) | [CAM 授权](https://cloud.tencent.com/document/product/862/117336)

**火山引擎：**
- [机器翻译文档](https://docs.volcengine.com/docs/4640/62099) | [文本翻译 API](https://docs.volcengine.com/docs/4640/65067) | [产品计费](https://docs.volcengine.com/docs/4640/68515)