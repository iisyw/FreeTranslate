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
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// TencentErrorResponse 腾讯云错误透传响应
type TencentErrorResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data TencentErrorData `json:"data"`
}

type TencentErrorData struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
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
	c.JSON(httpStatus, ErrorResponse{
		Code: code,
		Msg:  msg,
	})
}

// TencentErrorJSON 透传腾讯云错误
func TencentErrorJSON(c *gin.Context, httpStatus int, code int, codeStr string, message string, requestId string) {
	c.JSON(httpStatus, TencentErrorResponse{
		Code: code,
		Msg:  "translate error",
		Data: TencentErrorData{
			Code:      codeStr,
			Message:   message,
			RequestId: requestId,
		},
	})
}