package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/yeegeek/go-rest-api-starter/internal/config"
)

// JWTGenerator JWT 生成器接口
// 注意：本服务仅用于生成 JWT，不负责验证（验证由 API 网关完成）
type JWTGenerator interface {
	GenerateAccessToken(userID uint, email string, name string, roles []string) (string, error)
	GenerateRefreshToken(userID uint) (string, error)
}

type jwtGenerator struct {
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	db              *gorm.DB
}

// NewJWTGenerator 创建新的 JWT 生成器
func NewJWTGenerator(cfg *config.JWTConfig, db *gorm.DB) JWTGenerator {
	jwtSecret := cfg.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	accessTokenTTL := cfg.AccessTokenTTL
	if accessTokenTTL == 0 {
		if cfg.TTLHours > 0 {
			accessTokenTTL = time.Duration(cfg.TTLHours) * time.Hour
		} else {
			accessTokenTTL = 15 * time.Minute
		}
	}

	refreshTokenTTL := cfg.RefreshTokenTTL
	if refreshTokenTTL == 0 {
		refreshTokenTTL = 168 * time.Hour
	}

	return &jwtGenerator{
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		db:              db,
	}
}

// GenerateAccessToken 生成访问令牌
func (g *jwtGenerator) GenerateAccessToken(userID uint, email string, name string, roles []string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(g.accessTokenTTL)

	// 如果未提供角色，从数据库查询
	if roles == nil && g.db != nil {
		var roleNames []string
		err := g.db.Table("roles").
			Select("roles.name").
			Joins("JOIN user_roles ON user_roles.role_id = roles.id").
			Where("user_roles.user_id = ?", userID).
			Find(&roleNames).Error
		if err != nil {
			return "", fmt.Errorf("failed to fetch user roles: %w", err)
		}
		roles = roleNames
	}

	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", userID),
		"email": email,
		"name":  name,
		"roles": roles,
		"exp":   expirationTime.Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(g.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成刷新令牌
func (g *jwtGenerator) GenerateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	expirationTime := now.Add(g.refreshTokenTTL)

	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", userID),
		"type": "refresh",
		"exp":  expirationTime.Unix(),
		"iat":  now.Unix(),
		"nbf":  now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(g.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}
