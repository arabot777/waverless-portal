package handler

import (
	"net/http"

	"github.com/wavespeedai/waverless-portal/internal/service"

	"github.com/gin-gonic/gin"
)

type EndpointHandler struct {
	endpointService *service.EndpointService
	userService     *service.UserService
}

func NewEndpointHandler(endpointService *service.EndpointService, userService *service.UserService) *EndpointHandler {
	return &EndpointHandler{
		endpointService: endpointService,
		userService:     userService,
	}
}

// CreateEndpoint 创建 Endpoint
func (h *EndpointHandler) CreateEndpoint(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.GetString("org_id")

	var req service.CreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	endpoint, err := h.endpointService.Create(c.Request.Context(), userID, orgID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoint":       endpoint.LogicalName,
		"cluster":        endpoint.ClusterID,
		"price_per_hour": endpoint.PricePerHour,
		"status":         endpoint.Status,
	})
}

// ListEndpoints 列出用户的 Endpoints
func (h *EndpointHandler) ListEndpoints(c *gin.Context) {
	userID := c.GetString("user_id")

	endpoints, err := h.endpointService.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

// GetEndpoint 获取 Endpoint 详情 (透传 waverless)
func (h *EndpointHandler) GetEndpoint(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	// 从 waverless 获取实时详情
	detail, err := h.endpointService.GetEndpointDetail(c.Request.Context(), endpoint)
	if err != nil {
		// 如果 waverless 获取失败，返回本地数据
		c.JSON(http.StatusOK, endpoint)
		return
	}

	// 合并本地数据
	detail["logical_name"] = endpoint.LogicalName
	detail["price_per_hour"] = endpoint.PricePerHour
	detail["cluster_id"] = endpoint.ClusterID

	c.JSON(http.StatusOK, detail)
}

// UpdateEndpoint 更新 Endpoint 部署 (replicas, image, env)
func (h *EndpointHandler) UpdateEndpoint(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	var req struct {
		Replicas int               `json:"replicas"`
		Image    string            `json:"image"`
		Env      map[string]string `json:"env"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.endpointService.UpdateDeployment(c.Request.Context(), userID, name, req.Replicas, req.Image, req.Env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// UpdateEndpointConfig 更新 Endpoint 配置
func (h *EndpointHandler) UpdateEndpointConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.endpointService.UpdateConfig(c.Request.Context(), userID, name, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteEndpoint 删除 Endpoint
func (h *EndpointHandler) DeleteEndpoint(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	if err := h.endpointService.Delete(c.Request.Context(), userID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
