# FreeTranslate

基于腾讯云 MPS 的文本翻译 API 服务，支持多种语言。

## 功能特性

- 调用腾讯云 MPS `TextTranslation` 接口进行文本翻译
- 支持 `auto` 自动检测源语言
- 支持 200+ 种语言互译
- 单次请求最大 2000 Unicode 字符
- Bearer Token 鉴权
- Zap 日志记录（含字符数、计费时长）

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置

复制 `.env.example` 为 `.env`，填入实际值：

```bash
cp .env.example .env
```

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
| `text` | string | ✅    | 待翻译文本，最大 2000 Unicode 字符 |
| `source_lang` | string | ❌    | 源语言代码，默认 `auto`（自动检测） |
| `target_lang` | string | ✅    | 目标语言代码 |

**响应示例：**

```json
{
  "code": 10000,
  "msg": "success",
  "data": {
    "text": "你好，世界！",
    "source_lang": "en",
    "target_lang": "zh"
  }
}
```

**常用语言代码：**

| 代码 | 语言 | 代码 | 语言 |
|------|------|------|------|
| `zh` | 简体中文 | `en` | 英语 |
| `zh-TW` | 中文繁体 | `ja` | 日语 |
| `ko` | 韩语 | `fr` | 法语 |
| `de` | 德语 | `es` | 西班牙语 |
| `ru` | 俄语 | `ar` | 阿拉伯语 |
| `th` | 泰语 | `vi` | 越南语 |
| `yue` | 粤语 | `auto` | 自动检测 |

完整语言代码列表请参考 [腾讯云文档](https://cloud.tencent.com/document/product/862/126431)。

**错误响应示例：**

```json
// 参数缺失
{"code": 40001, "msg": "target_lang is required"}

// 文本超长
{"code": 42200, "msg": "text exceeds maximum length of 2000 characters"}

// 腾讯云错误透传
{"code": 50000, "msg": "translate error", "data": {"code": "UnsupportedOperation.TextTooLong", "message": "..."}}
```

---

## 腾讯云配置

### 开通 MPS 服务

1. 登录 [腾讯云控制台](https://console.cloud.tencent.com/)
2. 搜索 **媒体处理 MPS** → 进入产品页 → 点击 **立即开通**
3. 按页面提示完成开通

### 获取密钥

1. 进入 [访问密钥管理](https://console.cloud.tencent.com/cam/capi)
2. 创建密钥对（SecretId + SecretKey）
3. 将密钥填入 `.env`

### 子账户权限配置（CAM）

如果使用子账户的密钥，需要给子账户授予 MPS 访问权限：

#### 方式一：使用预设策略（推荐用于开发测试）

1. [CAM 控制台 → 用户](https://console.cloud.tencent.com/cam) → 选择你的子账户
2. 点击 **添加权限**
3. 搜索 `QcloudMPSFullAccess`，勾选并确认

#### 方式二：自定义最小权限策略（推荐用于生产环境）

1. 进入 [CAM 控制台 → 策略](https://console.cloud.tencent.com/cam/policy)
2. **创建策略** → **使用策略语法创建**
3. 粘贴以下策略内容：

```json
{
  "version": "2.0",
  "statement": [
    {
      "effect": "allow",
      "action": [
        "mps:TextTranslation"
      ],
      "resource": ["*"]
    }
  ]
}
```

4. 策略名称填写 `MPS-TextTranslation`，点击 **完成**
5. 进入 [CAM 控制台 → 用户](https://console.cloud.tencent.com/cam)，关联该策略到你的子账户

详细授权步骤请参考 [媒体处理账号授权文档](https://cloud.tencent.com/document/product/862/117336)。

---

## 计费说明

### 计费方式

腾讯云 MPS `TextTranslation` 使用 **翻译字幕（附加语种）** 计费项：

- **计费单位**：人民币（元）/ 分钟
- **换算规则**：每 **1100** Unicode 字符折算为 1 分钟
- 字符数按 Unicode 码点统计（例如：`hello` 算 5 字符，`你好` 算 2 字符）
- 按量计费价格：**0.2 元/分钟**（参考价格，以腾讯云实际定价为准）

> ⚠️ 注意：腾讯云 MPS TextTranslation 属于**机器翻译**产品线，与独立的机器翻译服务（TMT）共享账户余额和计费体系。

### 免费额度

腾讯云机器翻译服务为每个账号提供以下免费额度：

| 服务名称 | 免费额度 |
|----------|----------|
| 文本翻译 | **每月 500 万字符** |

免费额度按月发放，当月首次开通后立即发放当月额度，仅当月有效。扣费优先级：**免费额度 → 预付费资源包 → 后付费**。

详细免费额度说明：[腾讯云机器翻译计费概述](https://cloud.tencent.com/document/product/551/35017)。

### 计费计算示例

| 翻译文本 | Unicode 字符数 | 折算分钟 | 费用（元） |
|----------|---------------|----------|-----------|
| `Hello` | 5 | 0.0045 分钟 | ≈ 0.0009 元 |
| `你好，世界！` | 6 | 0.0055 分钟 | ≈ 0.0011 元 |
| 1000 字中文 | 1000 | 0.91 分钟 | ≈ 0.18 元 |
| 2000 字中文 | 2000 | 1.82 分钟 | ≈ 0.36 元 |

免费额度范围内（≤ 500万字符/月）**无需支付任何费用**。

更多定价详情：[腾讯云 MPS 按量计费说明](https://cloud.tencent.com/document/product/862/36180)

---

## 相关链接

- 腾讯云 MPS 官方文档：https://cloud.tencent.com/document/product/862
- 文本翻译 API 文档：https://cloud.tencent.com/document/product/862/126431
- 机器翻译计费概述：https://cloud.tencent.com/document/product/551/35017
- MPS 按量计费说明：https://cloud.tencent.com/document/product/862/36180
- CAM 授权文档：https://cloud.tencent.com/document/product/862/117336
- 腾讯云控制台：https://console.cloud.tencent.com/