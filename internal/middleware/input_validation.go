package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SQL 注入常见模式
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(\bunion\b.*\bselect\b)`),
	regexp.MustCompile(`(?i)(\bselect\b.*\bfrom\b)`),
	regexp.MustCompile(`(?i)(\binsert\b.*\binto\b)`),
	regexp.MustCompile(`(?i)(\bupdate\b.*\bset\b)`),
	regexp.MustCompile(`(?i)(\bdelete\b.*\bfrom\b)`),
	regexp.MustCompile(`(?i)(\bdrop\b.*\btable\b)`),
	regexp.MustCompile(`(?i)(\bexec\b|\bexecute\b)`),
	regexp.MustCompile(`(?i)(--|#|\/\*|\*\/)`), // SQL 注释
	regexp.MustCompile(`(?i)(\bor\b.*=.*)`),
	regexp.MustCompile(`(?i)(\band\b.*=.*)`),
	regexp.MustCompile(`(?i)(';|";)`), // 单引号或双引号结束
}

// XSS 常见模式
var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(<script[^>]*>.*?</script>)`),
	regexp.MustCompile(`(?i)(<iframe[^>]*>.*?</iframe>)`),
	regexp.MustCompile(`(?i)(javascript:)`),
	regexp.MustCompile(`(?i)(on\w+\s*=)`), // onclick, onerror 等事件
	regexp.MustCompile(`(?i)(<img[^>]*onerror[^>]*>)`),
}

// InputValidationMiddleware 输入验证中间件
// 检测并阻止常见的 SQL 注入和 XSS 攻击
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if isSuspicious(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid input detected in query parameter: " + key,
					})
					c.Abort()
					return
				}
			}
		}

		// 检查路径参数
		for _, param := range c.Params {
			if isSuspicious(param.Value) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid input detected in path parameter: " + param.Key,
				})
				c.Abort()
				return
			}
		}

		// 注意：请求体的验证应该在 Handler 层使用结构体验证标签完成
		// 这里主要关注 URL 参数的安全性

		c.Next()
	}
}

// isSuspicious 检查字符串是否包含可疑模式
func isSuspicious(input string) bool {
	// 检查 SQL 注入模式
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}

	// 检查 XSS 模式
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}

	// 检查是否包含多个连续的特殊字符（可能是攻击尝试）
	if strings.Contains(input, "''") || strings.Contains(input, "\"\"") {
		return true
	}

	return false
}

// SanitizeString 清理字符串，移除潜在的危险字符
// 注意：这不应该替代参数化查询，只是额外的防护层
func SanitizeString(input string) string {
	// 移除 SQL 注释
	input = strings.ReplaceAll(input, "--", "")
	input = strings.ReplaceAll(input, "/*", "")
	input = strings.ReplaceAll(input, "*/", "")
	input = strings.ReplaceAll(input, "#", "")

	// 移除多余的引号
	input = strings.ReplaceAll(input, "''", "'")
	input = strings.ReplaceAll(input, "\"\"", "\"")

	return strings.TrimSpace(input)
}
