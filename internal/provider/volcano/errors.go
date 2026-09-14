package volcano

import (
	"strings"

	"FreeTranslate/internal/provider"

	volcengineerr "github.com/volcengine/volcengine-go-sdk/volcengine/volcengineerr"
)

func classifyError(err error) error {
	if sdkErr, ok := err.(volcengineerr.Error); ok {
		requestID := ""
		if requestErr, ok := err.(volcengineerr.RequestFailure); ok {
			requestID = requestErr.RequestID()
		}
		return classifyCode(sdkErr.Code(), sdkErr.Message(), requestID, err)
	}
	return provider.NewTransportError(name, err)
}

func classifyCode(code, detail, requestID string, err error) error {
	codeLower := strings.ToLower(code)
	detailLower := strings.ToLower(detail)
	switch {
	case code == "FlowLimitExceeded":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorRateLimited, code, "服务商请求频率受限", detail, requestID, true, true, err)
	case code == "InternalServiceTimeout" || strings.Contains(codeLower, "timeout"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTimeout, code, "服务商响应超时", detail, requestID, true, true, err)
	case code == "ServiceUnavailable" || code == "ServiceUnavailableTemp" || code == "FailToConnect" || code == "InternalError":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnavailable, code, "服务商暂时不可用", detail, requestID, true, true, err)
	case code == "InvalidClientTokenId" || code == "SignatureDoesNotMatch" || code == "LackPolicy" || code == "AccessDenied" || code == "MissingAuthenticationToken":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnauthorized, code, "服务商鉴权或权限配置错误", detail, requestID, true, false, err)
	case code == "MissingParameter" || code == "MissingRequestInfo":
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorInvalidArgument, code, "服务商拒绝了请求参数", detail, requestID, false, false, err)
	case strings.Contains(detailLower, "unsupported") && strings.Contains(detailLower, "language"):
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorUnsupportedLanguage, code, "当前语言或语言对不受支持", detail, requestID, true, false, err)
	default:
		return provider.NewProviderErrorWithPolicy(name, provider.ErrorTranslationFailed, code, "火山引擎翻译服务处理失败", detail, requestID, true, true, err)
	}
}
