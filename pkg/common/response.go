package common

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 预定义错误
var (
	ParamBindError    = "参数绑定错误"
	UnauthorizedError = "未授权"
	ForbiddenError    = "权限不足"
	NotFoundError     = "资源不存在"
	ServerError       = "服务器内部错误"
)

// 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "操作成功",
		Data:    data,
	}
}

// 错误响应
func ErrorResponse(err interface{}) *Response {
	var message string
	switch v := err.(type) {
	case error:
		message = v.Error()
	case string:
		message = v
	default:
		message = ServerError
	}

	return &Response{
		Code:    500,
		Message: message,
	}
}
