package handler

import "github.com/gin-gonic/gin"

// 通用错误码 (0 - 1999)
const (
	SUCCESS             = 0    // 成功
	ERROR               = 1001 // 一般错误
	INVALID_PARAMS      = 1002 // 参数错误
	UNAUTHORIZED        = 1003 // 未授权
	NOT_FOUND           = 1004 // 资源不存在
	SERVER_ERROR        = 1005 // 服务器内部错误
	FORBIDDEN           = 1006 // 禁止访问
	TOO_MANY_REQUESTS   = 1007 // 请求过于频繁
	SERVICE_UNAVAILABLE = 1008 // 服务不可用
	TIMEOUT             = 1009 // 请求超时
	BAD_GATEWAY         = 1010 // 网关错误
)

// 用户相关错误码 (2000-2999)
const (
	USER_NOT_FOUND      = 2001 // 用户不存在
	USER_ALREADY_EXISTS = 2002 // 用户已存在
	INVALID_CREDENTIALS = 2003 // 无效的凭据
	ACCOUNT_DISABLED    = 2004 // 账户已禁用
	EMAIL_NOT_VERIFIED  = 2005 // 邮箱未验证
	PASSWORD_TOO_WEAK   = 2006 // 密码强度不够
	TOKEN_EXPIRED       = 2007 // 令牌已过期
	INVALID_TOKEN       = 2008 // 无效的令牌
)

// 数据相关错误码 (3000-3999)
const (
	DB_ERROR           = 3001 // 数据库错误
	TRANSACTION_FAILED = 3002 // 事务失败
	INVALID_QUERY      = 3003 // 无效的查询
	DATA_INTEGRITY     = 3004 // 数据完整性错误
	DUPLICATE_ENTRY    = 3005 // 重复数据
)

// 业务逻辑错误码 (4000-4999)
const (
	BUSINESS_ERROR    = 4001 // 业务逻辑错误
	VALIDATION_ERROR  = 4002 // 验证错误
	OPERATION_FAILED  = 4003 // 操作失败
	RESOURCE_LOCKED   = 4004 // 资源被锁定
	QUOTA_EXCEEDED    = 4005 // 超出配额
	RESOURCE_CONFLICT = 4006 // 资源冲突
)

// 第三方服务错误码 (5000-5999)
const (
	THIRD_PARTY_ERROR  = 5001 // 第三方服务错误
	API_LIMIT_EXCEEDED = 5002 // API调用次数超限
	SERVICE_DOWN       = 5003 // 第三方服务不可用
	INVALID_RESPONSE   = 5004 // 无效的响应
)

// 错误码对应的消息
var codeMessage = map[int]string{
	// 通用错误码
	SUCCESS:             "成功",
	ERROR:               "操作失败",
	INVALID_PARAMS:      "参数错误",
	UNAUTHORIZED:        "未授权",
	NOT_FOUND:           "资源不存在",
	SERVER_ERROR:        "服务器内部错误",
	FORBIDDEN:           "禁止访问",
	TOO_MANY_REQUESTS:   "请求过于频繁",
	SERVICE_UNAVAILABLE: "服务不可用",
	TIMEOUT:             "请求超时",
	BAD_GATEWAY:         "网关错误",

	// 用户相关错误码
	USER_NOT_FOUND:      "用户不存在",
	USER_ALREADY_EXISTS: "用户已存在",
	INVALID_CREDENTIALS: "无效的凭据",
	ACCOUNT_DISABLED:    "账户已禁用",
	EMAIL_NOT_VERIFIED:  "邮箱未验证",
	PASSWORD_TOO_WEAK:   "密码强度不够",
	TOKEN_EXPIRED:       "令牌已过期",
	INVALID_TOKEN:       "无效的令牌",

	// 数据相关错误码
	DB_ERROR:           "数据库错误",
	TRANSACTION_FAILED: "事务失败",
	INVALID_QUERY:      "无效的查询",
	DATA_INTEGRITY:     "数据完整性错误",
	DUPLICATE_ENTRY:    "重复数据",

	// 业务逻辑错误码
	BUSINESS_ERROR:    "业务逻辑错误",
	VALIDATION_ERROR:  "验证错误",
	OPERATION_FAILED:  "操作失败",
	RESOURCE_LOCKED:   "资源被锁定",
	QUOTA_EXCEEDED:    "超出配额",
	RESOURCE_CONFLICT: "资源冲突",

	// 第三方服务错误码
	THIRD_PARTY_ERROR:  "第三方服务错误",
	API_LIMIT_EXCEEDED: "API调用次数超限",
	SERVICE_DOWN:       "第三方服务不可用",
	INVALID_RESPONSE:   "无效的响应",
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
