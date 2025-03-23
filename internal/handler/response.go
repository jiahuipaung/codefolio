package handler

import "github.com/gin-gonic/gin"

// 错误码常量
const (
	SUCCESS        = 0    // 成功
	ERROR          = 1001 // 一般错误
	INVALID_PARAMS = 1002 // 参数错误
	UNAUTHORIZED   = 1003 // 未授权
	NOT_FOUND      = 1004 // 资源不存在
	SERVER_ERROR   = 1005 // 服务器内部错误
)

// 错误码对应的消息
var codeMessage = map[int]string{
	SUCCESS:        "成功",
	ERROR:          "操作失败",
	INVALID_PARAMS: "参数错误",
	UNAUTHORIZED:   "未授权",
	NOT_FOUND:      "资源不存在",
	SERVER_ERROR:   "服务器内部错误",
}

// Response 标准API响应结构
type Response struct {
	Code    int         `json:"code"`    // 错误码
	Message string      `json:"message"` // 错误信息
	Data    interface{} `json:"data"`    // 返回数据
}

// ResponseWithData 返回带数据的成功响应
func ResponseWithData(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    SUCCESS,
		Message: codeMessage[SUCCESS],
		Data:    data,
	})
}

// ResponseWithError 返回错误响应
func ResponseWithError(c *gin.Context, httpStatus int, errCode int, msg string) {
	message := msg
	if message == "" {
		message = codeMessage[errCode]
	}

	c.JSON(httpStatus, Response{
		Code:    errCode,
		Message: message,
		Data:    nil,
	})
}

// ResponseSuccess 返回简单的成功响应
func ResponseSuccess(c *gin.Context) {
	c.JSON(200, Response{
		Code:    SUCCESS,
		Message: codeMessage[SUCCESS],
		Data:    nil,
	})
}
