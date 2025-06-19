package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

// 系统级错误码
const (
	// 成功
	Success ErrorCode = 0

	// 系统错误 (1000-1999)
	ErrSystem             ErrorCode = 1000 // 系统错误
	ErrInternal           ErrorCode = 1001 // 内部错误
	ErrServiceUnavailable ErrorCode = 1002 // 服务不可用
	ErrDatabase           ErrorCode = 1003 // 数据库错误
	ErrCache              ErrorCode = 1004 // 缓存错误
	ErrConfig             ErrorCode = 1005 // 配置错误

	// 参数错误 (2000-2999)
	ErrParam         ErrorCode = 2000 // 参数错误
	ErrParamRequired ErrorCode = 2001 // 参数必填
	ErrParamInvalid  ErrorCode = 2002 // 参数无效
	ErrParamType     ErrorCode = 2003 // 参数类型错误
	ErrParamFormat   ErrorCode = 2004 // 参数格式错误

	// 认证错误 (3000-3999)
	ErrAuth         ErrorCode = 3000 // 认证错误
	ErrToken        ErrorCode = 3001 // Token错误
	ErrTokenExpired ErrorCode = 3002 // Token过期
	ErrPermission   ErrorCode = 3003 // 权限错误
	ErrRole         ErrorCode = 3004 // 角色错误

	// 业务错误 (4000-4999)
	ErrBusiness  ErrorCode = 4000 // 业务错误
	ErrNotFound  ErrorCode = 4001 // 资源不存在
	ErrDuplicate ErrorCode = 4002 // 资源重复
	ErrStatus    ErrorCode = 4003 // 状态错误
	ErrOperation ErrorCode = 4004 // 操作错误
)

// Error 自定义错误结构
type Error struct {
	Code    ErrorCode `json:"code"`              // 错误码
	Message string    `json:"message"`           // 错误信息
	Details string    `json:"details,omitempty"` // 错误详情
	Err     error     `json:"-"`                 // 原始错误
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *Error) HTTPStatus() int {
	switch {
	case e.Code >= 1000 && e.Code < 2000:
		return http.StatusInternalServerError
	case e.Code >= 2000 && e.Code < 3000:
		return http.StatusBadRequest
	case e.Code >= 3000 && e.Code < 4000:
		return http.StatusUnauthorized
	case e.Code >= 4000 && e.Code < 5000:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// New 创建新的错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装已有错误
func Wrap(err error, code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails 添加错误详情
func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

// 错误码映射表
var errorMessages = map[ErrorCode]string{
	Success:               "成功",
	ErrSystem:             "系统错误",
	ErrInternal:           "内部错误",
	ErrServiceUnavailable: "服务不可用",
	ErrDatabase:           "数据库错误",
	ErrCache:              "缓存错误",
	ErrConfig:             "配置错误",
	ErrParam:              "参数错误",
	ErrParamRequired:      "参数必填",
	ErrParamInvalid:       "参数无效",
	ErrParamType:          "参数类型错误",
	ErrParamFormat:        "参数格式错误",
	ErrAuth:               "认证错误",
	ErrToken:              "Token错误",
	ErrTokenExpired:       "Token过期",
	ErrPermission:         "权限错误",
	ErrRole:               "角色错误",
	ErrBusiness:           "业务错误",
	ErrNotFound:           "资源不存在",
	ErrDuplicate:          "资源重复",
	ErrStatus:             "状态错误",
	ErrOperation:          "操作错误",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// RegisterMessage 注册错误码消息
func RegisterMessage(code ErrorCode, message string) {
	errorMessages[code] = message
}
