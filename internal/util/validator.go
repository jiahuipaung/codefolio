package util

import (
	"codefolio/internal/handler"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 自定义验证错误信息
var validationErrorMessages = map[string]string{
	"required": "该字段是必填项",
	"email":    "请输入有效的电子邮件地址",
	"min":      "该字段长度不足",
	"max":      "该字段长度超出限制",
	"alphanum": "该字段只能包含字母和数字",
	"numeric":  "该字段只能包含数字",
	"eqfield":  "两个字段必须匹配",
}

// BindAndValidate 绑定并验证请求参数
func BindAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		// 处理验证错误
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := formatValidationErrors(validationErrors)
			handler.ResponseWithError(c, http.StatusBadRequest, handler.INVALID_PARAMS, strings.Join(errorMessages, "; "))
		} else {
			// 其他绑定错误
			handler.ResponseWithError(c, http.StatusBadRequest, handler.INVALID_PARAMS, "无效的请求参数格式")
		}
		return false
	}
	return true
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(errors validator.ValidationErrors) []string {
	var result []string

	for _, err := range errors {
		field := toSnakeCase(err.Field())
		tag := err.Tag()

		// 查找自定义错误消息
		message, exists := validationErrorMessages[tag]
		if !exists {
			message = "验证失败"
		}

		// 对特定标签添加额外信息
		switch tag {
		case "min", "max":
			message = message + " (" + err.Param() + ")"
		}

		result = append(result, field+": "+message)
	}

	return result
}

// toSnakeCase 转换为蛇形命名
func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
