package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义通用错误代码
const (
	// 通用错误代码 (1000-1999)
	CodeSuccess          = 0    // 成功
	CodeUnknownError     = 1000 // 未知错误
	CodeInternalError    = 1001 // 内部服务器错误
	CodeInvalidParams    = 1002 // 无效的参数
	CodeUnauthorized     = 1003 // 未授权
	CodeForbidden        = 1004 // 禁止访问
	CodeNotFound         = 1005 // 资源不存在
	CodeMethodNotAllowed = 1006 // 方法不允许
	CodeTimeout          = 1007 // 请求超时
	CodeBadRequest       = 1008 // 错误的请求
	CodeValidationFailed = 1009 // 验证失败
	CodeRequestFailed    = 1010 // 请求失败

	// 用户相关错误代码 (2000-2999)
	CodeUserNotFound          = 2000 // 用户不存在
	CodeUserAlreadyExists     = 2001 // 用户已存在
	CodeInvalidCredentials    = 2002 // 无效的凭据
	CodePasswordMismatch      = 2003 // 密码不匹配
	CodeInvalidToken          = 2004 // 无效的令牌
	CodeTokenExpired          = 2005 // 令牌已过期
	CodeEmailNotVerified      = 2006 // 邮箱未验证
	CodeInvalidEmail          = 2007 // 无效的邮箱
	CodeInvalidPassword       = 2008 // 无效的密码
	CodePasswordTooWeak       = 2009 // 密码强度不足
	CodeTooManyRequests       = 2010 // 请求次数过多
	CodeAccountLocked         = 2011 // 账户已锁定
	CodeSessionExpired        = 2012 // 会话已过期
	CodeLoginRequired         = 2013 // 需要登录
	CodePermissionDenied      = 2014 // 权限不足
	CodeUserDisabled          = 2015 // 用户已禁用
	CodeUserProfileIncomplete = 2016 // 用户资料不完整

	// 数据相关错误代码 (3000-3999)
	CodeDataNotFound         = 3000 // 数据不存在
	CodeDataAlreadyExists    = 3001 // 数据已存在
	CodeDataInvalid          = 3002 // 数据无效
	CodeDataCreateFailed     = 3003 // 创建数据失败
	CodeDataUpdateFailed     = 3004 // 更新数据失败
	CodeDataDeleteFailed     = 3005 // 删除数据失败
	CodeDatabaseError        = 3006 // 数据库错误
	CodeDataQueryFailed      = 3007 // 查询数据失败
	CodeDataSaveFailed       = 3008 // 保存数据失败
	CodeDataParsingFailed    = 3009 // 解析数据失败
	CodeDataValidationFailed = 3010 // 数据验证失败
	CodeDataTooLarge         = 3011 // 数据太大
	CodeDataCorrupted        = 3012 // 数据已损坏
	CodeDataVersionMismatch  = 3013 // 数据版本不匹配
	CodeDataLockFailed       = 3014 // 数据锁定失败
	CodeDataRelationInvalid  = 3015 // 数据关系无效

	// 业务逻辑错误代码 (4000-4999)
	CodeOperationFailed     = 4000 // 操作失败
	CodeOperationNotAllowed = 4001 // 操作不允许
	CodeOperationTimeout    = 4002 // 操作超时
	CodeResourceExhausted   = 4003 // 资源耗尽
	CodeQuotaExceeded       = 4004 // 配额超出
	CodeInvalidState        = 4005 // 无效的状态
	CodeInvalidAction       = 4006 // 无效的操作
	CodeFeatureDisabled     = 4007 // 功能已禁用
	CodeBillingError        = 4008 // 计费错误
	CodePaymentRequired     = 4009 // 需要付款
	CodeInsufficientBalance = 4010 // 余额不足
	CodeLimitExceeded       = 4011 // 超出限制
	CodeRateLimited         = 4012 // 速率受限
	CodeInsufficientStorage = 4013 // 存储空间不足
	CodeBusinessRuleFailed  = 4014 // 业务规则失败
	CodeWorkflowFailed      = 4015 // 工作流失败

	// 第三方服务错误代码 (5000-5999)
	CodeThirdPartyError      = 5000 // 第三方服务错误
	CodeRemoteServiceError   = 5001 // 远程服务错误
	CodeAPIError             = 5002 // API错误
	CodeNetworkError         = 5003 // 网络错误
	CodeGatewayError         = 5004 // 网关错误
	CodeServiceUnavailable   = 5005 // 服务不可用
	CodeDependencyFailed     = 5006 // 依赖项失败
	CodeExternalTimeout      = 5007 // 外部服务超时
	CodeAuthProviderError    = 5008 // 认证提供者错误
	CodeStorageProviderError = 5009 // 存储提供者错误
	CodeNotificationError    = 5010 // 通知错误
	CodeWebhookError         = 5011 // Webhook错误
	CodeIntegrationError     = 5012 // 集成错误
)

// 错误码到错误消息的映射
var codeMessages = map[int]string{
	// 通用错误代码 (1000-1999)
	CodeSuccess:          "操作成功",
	CodeUnknownError:     "未知错误",
	CodeInternalError:    "内部服务器错误",
	CodeInvalidParams:    "无效的参数",
	CodeUnauthorized:     "未授权",
	CodeForbidden:        "禁止访问",
	CodeNotFound:         "资源不存在",
	CodeMethodNotAllowed: "方法不允许",
	CodeTimeout:          "请求超时",
	CodeBadRequest:       "错误的请求",
	CodeValidationFailed: "验证失败",
	CodeRequestFailed:    "请求失败",

	// 用户相关错误代码 (2000-2999)
	CodeUserNotFound:          "用户不存在",
	CodeUserAlreadyExists:     "用户已存在",
	CodeInvalidCredentials:    "无效的凭据",
	CodePasswordMismatch:      "密码不匹配",
	CodeInvalidToken:          "无效的令牌",
	CodeTokenExpired:          "令牌已过期",
	CodeEmailNotVerified:      "邮箱未验证",
	CodeInvalidEmail:          "无效的邮箱",
	CodeInvalidPassword:       "无效的密码",
	CodePasswordTooWeak:       "密码强度不足",
	CodeTooManyRequests:       "请求次数过多",
	CodeAccountLocked:         "账户已锁定",
	CodeSessionExpired:        "会话已过期",
	CodeLoginRequired:         "需要登录",
	CodePermissionDenied:      "权限不足",
	CodeUserDisabled:          "用户已禁用",
	CodeUserProfileIncomplete: "用户资料不完整",

	// 数据相关错误代码 (3000-3999)
	CodeDataNotFound:         "数据不存在",
	CodeDataAlreadyExists:    "数据已存在",
	CodeDataInvalid:          "数据无效",
	CodeDataCreateFailed:     "创建数据失败",
	CodeDataUpdateFailed:     "更新数据失败",
	CodeDataDeleteFailed:     "删除数据失败",
	CodeDatabaseError:        "数据库错误",
	CodeDataQueryFailed:      "查询数据失败",
	CodeDataSaveFailed:       "保存数据失败",
	CodeDataParsingFailed:    "解析数据失败",
	CodeDataValidationFailed: "数据验证失败",
	CodeDataTooLarge:         "数据太大",
	CodeDataCorrupted:        "数据已损坏",
	CodeDataVersionMismatch:  "数据版本不匹配",
	CodeDataLockFailed:       "数据锁定失败",
	CodeDataRelationInvalid:  "数据关系无效",

	// 业务逻辑错误代码 (4000-4999)
	CodeOperationFailed:     "操作失败",
	CodeOperationNotAllowed: "操作不允许",
	CodeOperationTimeout:    "操作超时",
	CodeResourceExhausted:   "资源耗尽",
	CodeQuotaExceeded:       "配额超出",
	CodeInvalidState:        "无效的状态",
	CodeInvalidAction:       "无效的操作",
	CodeFeatureDisabled:     "功能已禁用",
	CodeBillingError:        "计费错误",
	CodePaymentRequired:     "需要付款",
	CodeInsufficientBalance: "余额不足",
	CodeLimitExceeded:       "超出限制",
	CodeRateLimited:         "速率受限",
	CodeInsufficientStorage: "存储空间不足",
	CodeBusinessRuleFailed:  "业务规则失败",
	CodeWorkflowFailed:      "工作流失败",

	// 第三方服务错误代码 (5000-5999)
	CodeThirdPartyError:      "第三方服务错误",
	CodeRemoteServiceError:   "远程服务错误",
	CodeAPIError:             "API错误",
	CodeNetworkError:         "网络错误",
	CodeGatewayError:         "网关错误",
	CodeServiceUnavailable:   "服务不可用",
	CodeDependencyFailed:     "依赖项失败",
	CodeExternalTimeout:      "外部服务超时",
	CodeAuthProviderError:    "认证提供者错误",
	CodeStorageProviderError: "存储提供者错误",
	CodeNotificationError:    "通知错误",
	CodeWebhookError:         "Webhook错误",
	CodeIntegrationError:     "集成错误",
}

// 获取错误码对应的消息
func GetMessage(code int) string {
	msg, ok := codeMessages[code]
	if !ok {
		return "未知错误"
	}
	return msg
}

// Response 定义统一的API响应格式
type Response struct {
	Code    int         `json:"code"`    // 错误码
	Message string      `json:"message"` // 错误消息
	Data    interface{} `json:"data"`    // 响应数据
}

// ResponseWithData 返回包含数据的成功响应
func ResponseWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: GetMessage(CodeSuccess),
		Data:    data,
	})
}

// ResponseWithError 返回错误响应
func ResponseWithError(c *gin.Context, code int, statusCode ...int) {
	httpStatus := http.StatusOK
	if len(statusCode) > 0 {
		httpStatus = statusCode[0]
	}

	c.JSON(httpStatus, Response{
		Code:    code,
		Message: GetMessage(code),
		Data:    nil,
	})
}

// ResponseSuccess 返回简单的成功响应
func ResponseSuccess(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: GetMessage(CodeSuccess),
		Data:    nil,
	})
}

// ResponseWithCustomError 返回自定义错误消息的错误响应
func ResponseWithCustomError(c *gin.Context, code int, message string, statusCode ...int) {
	httpStatus := http.StatusOK
	if len(statusCode) > 0 {
		httpStatus = statusCode[0]
	}

	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
