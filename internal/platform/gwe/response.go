package gwe

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code      int    `json:"code"`
	Type      string `json:"type,omitempty"`
	Msg       string `json:"msg"`
	Provider  string `json:"provider,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// SuccessJSON 发送成功的 JSON 响应
func SuccessJSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 10000,
		Msg:  "success",
		Data: data,
	})
}

// ErrorJSON 发送错误的 JSON 响应
func ErrorJSON(c *gin.Context, httpStatus int, code int, msg string) {
	ErrorJSONWithType(c, httpStatus, code, "", msg, "", false, "")
}

// ErrorJSONWithType sends a stable error category and provider metadata.
func ErrorJSONWithType(c *gin.Context, httpStatus int, code int, errorType, msg, providerName string, retryable bool, requestID string) {
	c.JSON(httpStatus, ErrorResponse{
		Code:      code,
		Type:      errorType,
		Msg:       msg,
		Provider:  providerName,
		Retryable: retryable,
		RequestID: requestID,
	})
}
