package handler

import (
	"net/http"

	"github.com/wavespeedai/waverless-portal/app/middleware"
	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/wavespeed"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetCurrentUser 获取当前用户信息
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	cookieName := config.GetCookieName()
	tokenString, _ := c.Cookie(cookieName)
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	// 从主站获取用户信息
	user, err := wavespeed.GetUserInfo(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to verify user"})
		return
	}

	// 从 JWT claims 获取 org 信息
	claims, exists := c.Get("claims")
	var orgID, role string
	if exists {
		if cl, ok := claims.(*middleware.Claims); ok {
			orgID = cl.OrgID
			role = cl.Role
		}
	}

	// 检查是否为管理员
	userType := "regular"
	if config.IsAdminEmail(user.Email) {
		role = "admin"
		userType = "admin"
	}

	// 获取用户余额 (从主站)
	var balance int64
	if orgID != "" {
		balance, _ = wavespeed.GetOrgBalance(c.Request.Context(), orgID, tokenString)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     user.UserID,
		"org_id":      orgID,
		"user_name":   user.Name,
		"email":       user.Email,
		"avatar_url":  user.AvatarURL,
		"role":        role,
		"user_type":   userType,
		"permissions": user.Permissions,
		"balance":     ToUSD(balance),
	})
}
