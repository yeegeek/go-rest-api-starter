package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// HeaderUserID 用户 ID 头
	HeaderUserID = "X-User-ID"
	// HeaderUserRole 用户角色头
	HeaderUserRole = "X-User-Role"
	// ContextKeyUserID 上下文中的用户 ID 键
	ContextKeyUserID = "user_id"
	// ContextKeyUserRole 上下文中的用户角色键
	ContextKeyUserRole = "user_role"
)

// GatewayAuthMiddleware 网关认证中间件
// 从网关传递的 HTTP 头中提取用户信息
func GatewayAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(HeaderUserID)
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing user ID header",
			})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user ID format",
			})
			c.Abort()
			return
		}

		userRole := c.GetHeader(HeaderUserRole)
		if userRole == "" {
			userRole = "user" // 默认角色
		}

		// 将用户信息存储到上下文
		c.Set(ContextKeyUserID, uint(userID))
		c.Set(ContextKeyUserRole, userRole)

		c.Next()
	}
}

// GetUserIDFromContext 从上下文获取用户 ID
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetUserRoleFromContext 从上下文获取用户角色
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyUserRole)
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	return roleStr, ok
}

// RequireRole 要求特定角色的中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := GetUserRoleFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user role not found",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		hasRole := false
		for _, role := range roles {
			if strings.EqualFold(userRole, role) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminRole 要求管理员角色的中间件
func RequireAdminRole() gin.HandlerFunc {
	return RequireRole("admin")
}
