package util

import (
	"codefolio/internal/common"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 自定义验证错误消息映射
var validationErrorMessages = map[string]string{
	"required": "必须填写",
	"email":    "必须是有效的电子邮件地址",
	"min":      "长度不能小于 %s",
	"max":      "长度不能大于 %s",
	"len":      "长度必须是 %s",
	"eq":       "必须等于 %s",
	"ne":       "不能等于 %s",
	"gt":       "必须大于 %s",
	"gte":      "必须大于等于 %s",
	"lt":       "必须小于 %s",
	"lte":      "必须小于等于 %s",
}

// BindAndValidate 绑定并验证请求参数
// 返回错误，如果有错误则在函数内部已经处理了HTTP响应
func BindAndValidate(c *gin.Context, obj interface{}) error {
	// 绑定请求参数
	if err := c.ShouldBindJSON(obj); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// 处理验证错误
			fieldErrors := formatValidationErrors(validationErrors)
			common.ResponseWithError(c, common.CodeInvalidParams, http.StatusBadRequest)
			return err
		}
		// 其他绑定错误
		common.ResponseWithError(c, common.CodeBadRequest, http.StatusBadRequest)
		return err
	}
	return nil
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(errors validator.ValidationErrors) map[string]string {
	errorMap := make(map[string]string)

	for _, err := range errors {
		field := strings.ToLower(err.Field())
		tag := err.Tag()

		message, ok := validationErrorMessages[tag]
		if !ok {
			message = fmt.Sprintf("验证失败: %s", tag)
		} else if strings.Contains(message, "%s") {
			message = fmt.Sprintf(message, err.Param())
		}

		errorMap[field] = message
	}

	return errorMap
}

// toSnakeCase 转换为蛇形命名
func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
