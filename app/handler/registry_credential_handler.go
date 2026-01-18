package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"github.com/wavespeedai/waverless-portal/pkg/utils"
)

type RegistryCredentialHandler struct {
	repo *mysql.RegistryCredentialRepo
}

func NewRegistryCredentialHandler(repo *mysql.RegistryCredentialRepo) *RegistryCredentialHandler {
	return &RegistryCredentialHandler{repo: repo}
}

// Create 创建凭证
func (h *RegistryCredentialHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.GetString("org_id")

	var req struct {
		Name     string `json:"name" binding:"required"`
		Registry string `json:"registry"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Registry == "" {
		req.Registry = "docker.io"
	}

	// 加密密码
	encrypted, err := utils.Encrypt(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt password"})
		return
	}

	cred := &model.RegistryCredential{
		UserID:            userID,
		OrgID:             orgID,
		Name:              req.Name,
		Registry:          req.Registry,
		Username:          req.Username,
		PasswordEncrypted: encrypted,
	}

	if err := h.repo.Create(c.Request.Context(), cred); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       cred.ID,
		"name":     cred.Name,
		"registry": cred.Registry,
		"username": cred.Username,
	})
}

// List 列出凭证
func (h *RegistryCredentialHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	creds, err := h.repo.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 不返回密码
	result := make([]gin.H, len(creds))
	for i, cred := range creds {
		result[i] = gin.H{
			"id":         cred.ID,
			"name":       cred.Name,
			"registry":   cred.Registry,
			"username":   cred.Username,
			"created_at": cred.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"credentials": result})
}

// Delete 删除凭证
func (h *RegistryCredentialHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	if err := h.repo.Delete(c.Request.Context(), userID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
