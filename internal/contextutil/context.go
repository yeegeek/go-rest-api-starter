package contextutil

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Context keys
const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
)

// GetUserID 从上下文获取用户 ID
// 返回 0 表示未找到
func GetUserID(c *gin.Context) uint {
	userIDValue, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	
	switch v := userIDValue.(type) {
	case uint:
		return v
	case int:
		return uint(v)
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0
		}
		return uint(id)
	default:
		return 0
	}
}

// MustGetUserID 获取用户 ID 或返回错误
func MustGetUserID(c *gin.Context) (uint, error) {
	userID := GetUserID(c)
	if userID == 0 {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetUserRole 从上下文获取用户角色
// 返回空字符串表示未找到
func GetUserRole(c *gin.Context) string {
	roleValue, exists := c.Get(UserRoleKey)
	if !exists {
		return ""
	}
	
	if role, ok := roleValue.(string); ok {
		return role
	}
	return ""
}

// MustGetUserRole 获取用户角色或返回错误
func MustGetUserRole(c *gin.Context) (string, error) {
	role := GetUserRole(c)
	if role == "" {
		return "", fmt.Errorf("user role not found in context")
	}
	return role, nil
}

// IsAuthenticated 检查请求是否已认证
func IsAuthenticated(c *gin.Context) bool {
	return GetUserID(c) != 0
}

// CanAccessUser 检查认证用户是否可以访问目标用户
func CanAccessUser(c *gin.Context, targetUserID uint) bool {
	if IsAdmin(c) {
		return true
	}
	authenticatedUserID := GetUserID(c)
	return authenticatedUserID == targetUserID
}

// HasRole 检查用户是否具有特定角色
func HasRole(c *gin.Context, role string) bool {
	userRole := GetUserRole(c)
	return userRole == role
}

// IsAdmin 检查用户是否是管理员
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, "admin")
}

// GetRoles 获取用户角色列表（网关模式下只有一个角色）
func GetRoles(c *gin.Context) []string {
	role := GetUserRole(c)
	if role == "" {
		return []string{}
	}
	return []string{role}
}

// SetUserID 设置用户 ID 到上下文
func SetUserID(c *gin.Context, userID uint) {
	c.Set(UserIDKey, userID)
}

// SetUserRole 设置用户角色到上下文
func SetUserRole(c *gin.Context, role string) {
	c.Set(UserRoleKey, role)
}
