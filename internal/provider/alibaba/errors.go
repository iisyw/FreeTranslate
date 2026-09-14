package alibaba

import (
	"strings"

	"FreeTranslate/internal/provider"

	"github.com/alibabacloud-go/tea/tea"
)

func classifyError(err error) error {
	if sdkErr, ok := err.(*tea.SDKError); ok {
		return classifyCode(tea.StringValue(sdkErr.Code), tea.StringValue(sdkErr.Message), "", err)
	}
	return provider.NewTransportError(name, err)
}

func classifyCode(code, detail, requestID string, err error) error {
	switch strings.TrimSpace(code) {
	case "10001":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTimeout, code, "服务商响应超时", detail, requestID, true, true, err)
	case "10002", "10006", "10007", "10010", "10011", "10012", "10013", "19999":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnavailable, code, "服务商暂时不可用或免费额度已用尽", detail, requestID, true, true, err)
	case "10003", "10004":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorInvalidArgument, code, "服务商拒绝了请求参数", detail, requestID, false, false, err)
	case "10005":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnsupportedLanguage, code, "当前语言或语言对不受支持", detail, requestID, true, false, err)
	case "10008":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTextTooLong, code, "文本长度超过服务商限制", detail, requestID, true, false, err)
	case "10009":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnauthorized, code, "服务商鉴权或权限配置错误", detail, requestID, true, false, err)
	default:
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTranslationFailed, code, "阿里云翻译服务处理失败", detail, requestID, true, true, err)
	}
}
