package middleware

import (
	"net/http"
	"strings"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/wavespeed"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT Claims
type Claims struct {
	UserID      string   `json:"user_id"`
	OrgID       string   `json:"org_id"`
	UserName    string   `json:"user_name"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// JWTAuthWithUserService JWT 认证中间件（自动创建用户余额）
func JWTAuthWithUserService(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieName := config.GetCookieName()
		tokenString, err := c.Cookie(cookieName)
		if err != nil || tokenString == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				tokenString = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GlobalConfig.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("org_id", claims.OrgID)
		c.Set("email", claims.Email)
		c.Set("user_name", claims.UserName)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		// 自动创建/更新用户记录
		if userService != nil {
			userService.EnsureUser(c.Request.Context(), claims.UserID, claims.OrgID, claims.UserName, claims.Email)
		}

		c.Next()
	}
}

// JWTAuth JWT 认证中间件（不自动创建用户）
func JWTAuth() gin.HandlerFunc {
	return JWTAuthWithUserService(nil)
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !config.IsAdminEmail(email.(string)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyOrJWTAuth 支持 API Key 或 JWT 认证
func APIKeyOrJWTAuth(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先尝试 API Key (Authorization: Bearer xxx)
		auth := c.GetHeader("Authorization")
		if auth != "" {
			// API Key，尝试验证
			info, err := wavespeed.ValidateAPIKey(c.Request.Context(), auth)
			if err == nil {
				c.Set("user_id", info.UserID)
				c.Set("org_id", info.OrgID)
				c.Set("email", info.Email)
				c.Set("auth_type", "apikey")
				if userService != nil {
					userService.EnsureUser(c.Request.Context(), info.UserID, info.OrgID, "", info.Email)
				}
				c.Next()
				return
			}
		}

		// 回退到 JWT 认证
		cookieName := config.GetCookieName()
		tokenString, _ := c.Cookie(cookieName)
		if tokenString == "" && strings.HasPrefix(auth, "Bearer ") {
			tokenString = strings.TrimPrefix(auth, "Bearer ")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GlobalConfig.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("org_id", claims.OrgID)
		c.Set("email", claims.Email)
		c.Set("user_name", claims.UserName)
		c.Set("role", claims.Role)
		c.Set("claims", claims)
		c.Set("auth_type", "jwt")

		if userService != nil {
			userService.EnsureUser(c.Request.Context(), claims.UserID, claims.OrgID, claims.UserName, claims.Email)
		}

		c.Next()
	}
}
