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

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string) *Response {
	return &Response{
		Code:    -1,
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
