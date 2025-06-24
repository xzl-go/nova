package nova

// Error 错误类型
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewError 创建错误
func NewError(code int, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Message
}
