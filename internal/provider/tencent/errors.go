package tencent

import (
	"strings"

	"FreeTranslate/internal/provider"

	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

func classifyError(err error) error {
	if sdkErr, ok := err.(*tcerr.TencentCloudSDKError); ok {
		return classifyCode(sdkErr.Code, sdkErr.Message, sdkErr.RequestId, err)
	}
	return provider.NewTransportError(name, err)
}

func classifyCode(code, detail, requestID string, err error) error {
	codeLower := strings.ToLower(code)
	switch {
	case strings.Contains(codeLower, "texttoolong"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTextTooLong, code, "文本长度超过服务商限制", detail, requestID, true, false, err)
	case strings.Contains(codeLower, "unsupportedlanguage"),
		strings.Contains(codeLower, "unsupportedtargetlanguage"),
		strings.Contains(codeLower, "unsupportedsourcelanguage"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnsupportedLanguage, code, "当前语言或语言对不受支持", detail, requestID, true, false, err)
	case strings.Contains(codeLower, "limitexceeded"), strings.Contains(codeLower, "requestlimitexceeded"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorRateLimited, code, "服务商请求频率受限", detail, requestID, true, true, err)
	case strings.Contains(codeLower, "backendtimeout"), strings.Contains(codeLower, "timeout"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTimeout, code, "服务商响应超时", detail, requestID, true, true, err)
	case strings.Contains(codeLower, "nofreeamount"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnavailable, code, "服务商免费额度已用尽", detail, requestID, true, true, err)
	case strings.Contains(codeLower, "usernotregistered"), strings.Contains(codeLower, "serviceisolate"), strings.Contains(codeLower, "stopusing"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnavailable, code, "服务商未开通或暂时不可用", detail, requestID, true, false, err)
	case strings.Contains(codeLower, "unauthorized"), strings.Contains(codeLower, "authfailure"), strings.Contains(codeLower, "accessdenied"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnauthorized, code, "服务商鉴权或权限配置错误", detail, requestID, true, false, err)
	case strings.Contains(codeLower, "invalidparameter"), strings.Contains(codeLower, "missingparameter"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorInvalidArgument, code, "服务商拒绝了请求参数", detail, requestID, false, false, err)
	default:
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTranslationFailed, code, "腾讯云翻译服务处理失败", detail, requestID, true, true, err)
	}
}
