package handler

import (
	"net/http"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"

	"github.com/gin-gonic/gin"
)

type ClusterHandler struct {
	clusterService *service.ClusterService
}

func NewClusterHandler(clusterService *service.ClusterService) *ClusterHandler {
	return &ClusterHandler{clusterService: clusterService}
}

// CreateCluster 创建集群
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	var req struct {
		ClusterID   string `json:"cluster_id" binding:"required"`
		ClusterName string `json:"cluster_name" binding:"required"`
		Region      string `json:"region" binding:"required"`
		APIEndpoint string `json:"api_endpoint" binding:"required"`
		APIKey      string `json:"api_key"`
		Status      string `json:"status"`
		Priority    int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Priority == 0 {
		req.Priority = 100
	}
	if err := h.clusterService.CreateCluster(c.Request.Context(), req.ClusterID, req.ClusterName, req.Region, req.APIEndpoint, req.APIKey, req.Status, req.Priority); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "created"})
}

// UpdateCluster 更新集群
func (h *ClusterHandler) UpdateCluster(c *gin.Context) {
	clusterID := c.Param("id")
	var req struct {
		ClusterName string `json:"cluster_name"`
		Region      string `json:"region"`
		APIEndpoint string `json:"api_endpoint"`
		APIKey      string `json:"api_key"`
		Status      string `json:"status"`
		Priority    int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.clusterService.UpdateCluster(c.Request.Context(), clusterID, req.ClusterName, req.Region, req.APIEndpoint, req.APIKey, req.Status, req.Priority); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteCluster 删除集群
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	clusterID := c.Param("id")
	if err := h.clusterService.DeleteCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListClusters 列出所有集群
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	clusters, err := h.clusterService.ListClusters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clusters": clusters})
}

// GetCluster 获取集群详情
func (h *ClusterHandler) GetCluster(c *gin.Context) {
	clusterID := c.Param("id")

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

// GetClusterSpecs 获取集群规格
func (h *ClusterHandler) GetClusterSpecs(c *gin.Context) {
	clusterID := c.Param("id")
	specs, err := h.clusterService.GetClusterSpecs(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"specs": specs})
}

// CreateClusterSpec 创建集群规格
func (h *ClusterHandler) CreateClusterSpec(c *gin.Context) {
	clusterID := c.Param("id")
	var req struct {
		ClusterSpecName   string `json:"cluster_spec_name" binding:"required"`
		SpecName          string `json:"spec_name" binding:"required"`
		TotalCapacity     int    `json:"total_capacity"`
		AvailableCapacity int    `json:"available_capacity"`
		IsAvailable       *bool  `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}
	spec := &model.ClusterSpec{
		ClusterID: clusterID, ClusterSpecName: req.ClusterSpecName, SpecName: req.SpecName,
		TotalCapacity: req.TotalCapacity, AvailableCapacity: req.AvailableCapacity,
		IsAvailable: isAvailable,
	}
	if err := h.clusterService.CreateClusterSpec(c.Request.Context(), spec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "created", "spec": spec})
}

// UpdateClusterSpec 更新集群规格
func (h *ClusterHandler) UpdateClusterSpec(c *gin.Context) {
	var req struct {
		ID                int64 `json:"id" binding:"required"`
		TotalCapacity     int   `json:"total_capacity"`
		AvailableCapacity int   `json:"available_capacity"`
		IsAvailable       *bool `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.TotalCapacity > 0 {
		updates["total_capacity"] = req.TotalCapacity
	}
	if req.AvailableCapacity >= 0 {
		updates["available_capacity"] = req.AvailableCapacity
	}
	if req.IsAvailable != nil {
		updates["is_available"] = *req.IsAvailable
	}
	if err := h.clusterService.UpdateClusterSpec(c.Request.Context(), req.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteClusterSpec 删除集群规格
func (h *ClusterHandler) DeleteClusterSpec(c *gin.Context) {
	var req struct {
		ID int64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.clusterService.DeleteClusterSpec(c.Request.Context(), req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// WebhookHandler Webhook 处理器
type WebhookHandler struct {
	billingService *service.BillingService
}

func NewWebhookHandler(billingService *service.BillingService) *WebhookHandler {
	return &WebhookHandler{billingService: billingService}
}

// WorkerCreated Worker 创建通知 (deprecated - now using WorkerSyncJob)
func (h *WebhookHandler) WorkerCreated(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// WorkerTerminated Worker 终止通知 (deprecated - now using WorkerSyncJob)
func (h *WebhookHandler) WorkerTerminated(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
