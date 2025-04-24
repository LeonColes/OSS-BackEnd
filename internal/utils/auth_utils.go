package utils

import (
	"fmt"
	"strings"
)

// 操作类型常量
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// 资源类型常量
const (
	ResourceProject = "projects"
	ResourceGroup   = "groups"
	ResourceFile    = "files"
	ResourceUser    = "users"
	ResourceRole    = "roles"
)

// BuildUserSubject 构建用户主体标识
func BuildUserSubject(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

// BuildGroupDomain 构建群组域标识
func BuildGroupDomain(groupID string) string {
	return fmt.Sprintf("group:%s", groupID)
}

// BuildProjectDomain 构建项目域标识
func BuildProjectDomain(projectID string) string {
	return fmt.Sprintf("project:%s", projectID)
}

// MapMethodToAction 将HTTP方法映射为CRUD操作
func MapMethodToAction(method string) string {
	method = strings.ToUpper(method)
	switch method {
	case "POST":
		return ActionCreate
	case "GET":
		return ActionRead
	case "PUT", "PATCH":
		return ActionUpdate
	case "DELETE":
		return ActionDelete
	default:
		return ActionRead
	}
}

// ExtractIDFromSubject 从主体中提取ID
func ExtractIDFromSubject(subject string) (string, error) {
	parts := strings.Split(subject, ":")
	if len(parts) != 2 || parts[0] != "user" {
		return "", fmt.Errorf("无效的主体标识: %s", subject)
	}

	return parts[1], nil
}

// ExtractIDFromDomain 从域中提取ID
func ExtractIDFromDomain(domain string) (string, string, error) {
	parts := strings.Split(domain, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("无效的域标识: %s", domain)
	}

	return parts[1], parts[0], nil
}
