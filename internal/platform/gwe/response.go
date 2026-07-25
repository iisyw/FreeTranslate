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