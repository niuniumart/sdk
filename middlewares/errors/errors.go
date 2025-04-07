package errors

type Error interface {
	Error() string
	Code() int
	SetCode(code int)
	SetMessage(message string)
}

// 自定义错误类型
type sdkError struct {
	code    int    // 错误码
	message string // 错误消息
}

func (e *sdkError) Error() string {
	return e.message
}

func (e *sdkError) Code() int {
	return e.code
}

func (e *sdkError) SetCode(code int) {
	e.code = code
}

func (e *sdkError) SetMessage(message string) {
	e.message = message
}

// 创建一个新的自定义错误
func NewError(code int, message string) Error {
	return &sdkError{code: code, message: message}
}
