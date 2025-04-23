package utils

import (
	"net/http"
)

// MapMethodToAction HTTP方法映射到操作，用于Casbin权限检查
func MapMethodToAction(method string) string {
	switch method {
	case http.MethodGet:
		return "read"
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "*"
	}
}
