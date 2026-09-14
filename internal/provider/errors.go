package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// ErrorKind is the stable error category exposed by the API.
type ErrorKind string

const (
	ErrorInvalidArgument     ErrorKind = "INVALID_ARGUMENT"
	ErrorUnsupportedLanguage ErrorKind = "UNSUPPORTED_LANGUAGE"
	ErrorTextTooLong         ErrorKind = "TEXT_TOO_LONG"
	ErrorTimeout             ErrorKind = "PROVIDER_TIMEOUT"
	ErrorRateLimited         ErrorKind = "PROVIDER_RATE_LIMITED"
	ErrorUnavailable         ErrorKind = "PROVIDER_UNAVAILABLE"
	ErrorUnauthorized        ErrorKind = "PROVIDER_UNAUTHORIZED"
	ErrorTranslationFailed   ErrorKind = "TRANSLATION_FAILED"
	ErrorAllProvidersFailed  ErrorKind = "ALL_PROVIDERS_FAILED"
)

// ProviderError keeps provider details for logs while exposing a stable message.
type ProviderError struct {
	Kind         ErrorKind
	ProviderName string
	ProviderCode string
	Message      string
	Detail       string
	RequestID    string
	CanFailover  bool
	Retryable    bool
	Err          error
}

func (e *ProviderError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Detail != "" {
		return e.Detail
	}
	return string(e.Kind)
}

func (e *ProviderError) Unwrap() error { return e.Err }

func NewProviderErrorWithPolicy(providerName string, kind ErrorKind, code, message, detail, requestID string, canFailover, retryable bool, err error) *ProviderError {
	return &ProviderError{
		Kind:         kind,
		ProviderName: providerName,
		ProviderCode: code,
		Message:      message,
		Detail:       detail,
		RequestID:    requestID,
		CanFailover:  canFailover,
		Retryable:    retryable,
		Err:          err,
	}
}

func AsProviderError(err error) *ProviderError {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr
	}
	return nil
}

func CanFailover(err error) bool {
	if providerErr := AsProviderError(err); providerErr != nil {
		return providerErr.CanFailover
	}
	var netErr net.Error
	return errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary())
}

func NewTransportError(providerName string, err error) *ProviderError {
	if errors.Is(err, context.DeadlineExceeded) {
		return NewProviderErrorWithPolicy(providerName, ErrorTimeout, "", "服务商响应超时", err.Error(), "", true, true, err)
	}
	return NewProviderErrorWithPolicy(providerName, ErrorUnavailable, "", "服务商暂时不可用", err.Error(), "", true, true, err)
}

func NewTextTooLongError(providerName string, max int) *ProviderError {
	return NewProviderErrorWithPolicy(providerName, ErrorTextTooLong, "", fmt.Sprintf("文本长度超过服务商限制（最多 %d 个字符）", max), "", "", true, false, nil)
}
