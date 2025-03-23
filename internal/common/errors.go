package common

import (
	"errors"
)

// 常用错误常量
var (
	// ErrRecordNotFound 记录不存在错误
	ErrRecordNotFound = errors.New("记录不存在")

	// ErrInvalidInput 无效输入错误
	ErrInvalidInput = errors.New("无效的输入参数")

	// ErrInternalServer 内部服务器错误
	ErrInternalServer = errors.New("内部服务器错误")

	// ErrDuplicateRecord 记录重复错误
	ErrDuplicateRecord = errors.New("记录已存在")

	// ErrUnauthorized 未授权错误
	ErrUnauthorized = errors.New("未授权访问")

	// ErrForbidden 禁止访问错误
	ErrForbidden = errors.New("禁止访问")

	// ErrTimeout 超时错误
	ErrTimeout = errors.New("请求超时")

	// ErrNotImplemented 未实现错误
	ErrNotImplemented = errors.New("功能未实现")
)
