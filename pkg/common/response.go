package common

// Response API 统一响应结构
type Response struct {
	Code    int         `json:"code"`           // 状态码
	Message string      `json:"message"`        // 消息
	Data    interface{} `json:"data,omitempty"` // 数据
}

// 预定义错误
var (
	ParamBindError    = "参数绑定错误"
	UnauthorizedError = "未授权"
	ForbiddenError    = "权限不足"
	NotFoundError     = "资源不存在"
	ServerError       = "服务器内部错误"
)

// 预定义状态码
const (
	CodeSuccess      = 200 // 成功
	CodeError        = 400 // 错误
	CodeUnauthorized = 401 // 未授权
	CodeForbidden    = 403 // 禁止访问
	CodeNotFound     = 404 // 资源不存在
	CodeServerError  = 500 // 服务器错误
)

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string) *Response {
	return &Response{
		Code:    CodeError,
		Message: message,
	}
}

// ErrorWithCodeResponse 带状态码的错误响应
func ErrorWithCodeResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// UnauthorizedResponse 未授权响应
func UnauthorizedResponse() *Response {
	return &Response{
		Code:    CodeUnauthorized,
		Message: UnauthorizedError,
	}
}

// ForbiddenResponse 禁止访问响应
func ForbiddenResponse() *Response {
	return &Response{
		Code:    CodeForbidden,
		Message: ForbiddenError,
	}
}

// NotFoundResponse 资源不存在响应
func NotFoundResponse() *Response {
	return &Response{
		Code:    CodeNotFound,
		Message: NotFoundError,
	}
}

// ServerErrorResponse 服务器错误响应
func ServerErrorResponse() *Response {
	return &Response{
		Code:    CodeServerError,
		Message: ServerError,
	}
}
