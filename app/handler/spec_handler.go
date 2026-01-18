package handler

import (
	"net/http"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"

	"github.com/gin-gonic/gin"
)

type SpecHandler struct {
	specService *service.SpecService
}

func NewSpecHandler(specService *service.SpecService) *SpecHandler {
	return &SpecHandler{specService: specService}
}

// ListSpecs 列出所有可用规格（带集群可用性）
func (h *SpecHandler) ListSpecs(c *gin.Context) {
	specType := c.Query("type")
	specs, err := h.specService.ListSpecs(c.Request.Context(), specType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"specs": convertSpecs(specs)})
}

// ListAllSpecs 列出所有规格配置（管理用）
func (h *SpecHandler) ListAllSpecs(c *gin.Context) {
	specs, err := h.specService.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"specs": convertSpecPricings(specs)})
}

// GetSpec 获取规格详情
func (h *SpecHandler) GetSpec(c *gin.Context) {
	name := c.Param("name")
	spec, err := h.specService.GetSpec(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "spec not found"})
		return
	}
	c.JSON(http.StatusOK, convertSpecPricing(spec))
}

func convertSpecs(specs []mysql.SpecWithAvailability) []map[string]interface{} {
	result := make([]map[string]interface{}, len(specs))
	for i, s := range specs {
		result[i] = map[string]interface{}{
			"spec_name":        s.SpecName,
			"spec_type":        s.SpecType,
			"gpu_type":         s.GPUType,
			"gpu_count":        s.GPUCount,
			"cpu_cores":        s.CPUCores,
			"ram_gb":           s.RAMGB,
			"price_per_hour":   ToUSD(s.PricePerHour),
			"available_clusters": s.AvailableClusters,
		}
	}
	return result
}

func convertSpecPricings(specs []model.SpecPricing) []map[string]interface{} {
	result := make([]map[string]interface{}, len(specs))
	for i, s := range specs {
		result[i] = convertSpecPricing(&s)
	}
	return result
}

func convertSpecPricing(s *model.SpecPricing) map[string]interface{} {
	return map[string]interface{}{
		"id":             s.ID,
		"spec_name":      s.SpecName,
		"spec_type":      s.SpecType,
		"gpu_type":       s.GPUType,
		"gpu_count":      s.GPUCount,
		"cpu_cores":      s.CPUCores,
		"ram_gb":         s.RAMGB,
		"disk_gb":        s.DiskGB,
		"price_per_hour": ToUSD(s.PricePerHour),
		"min_price":      ToUSD(s.MinPrice),
		"max_price":      ToUSD(s.MaxPrice),
		"description":    s.Description,
		"is_available":   s.IsAvailable,
	}
}

// CreateSpec 创建规格
func (h *SpecHandler) CreateSpec(c *gin.Context) {
	var req model.SpecPricing
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.IsAvailable = true
	if err := h.specService.CreateSpec(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "created", "spec": req})
}

// UpdateSpec 更新规格
func (h *SpecHandler) UpdateSpec(c *gin.Context) {
	var req struct {
		ID                  int64   `json:"id" binding:"required"`
		SpecName     string  `json:"spec_name"`
		SpecType     string  `json:"spec_type"`
		GPUType      string  `json:"gpu_type"`
		GPUCount     int     `json:"gpu_count"`
		CPUCores     int     `json:"cpu_cores"`
		RAMGB        int     `json:"ram_gb"`
		DiskGB       int     `json:"disk_gb"`
		PricePerHour float64 `json:"price_per_hour"`
		Description  string  `json:"description"`
		IsAvailable  *bool   `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.SpecName != "" { updates["spec_name"] = req.SpecName }
	if req.SpecType != "" { updates["spec_type"] = req.SpecType }
	if req.GPUType != "" { updates["gpu_type"] = req.GPUType }
	if req.GPUCount > 0 { updates["gpu_count"] = req.GPUCount }
	if req.CPUCores > 0 { updates["cpu_cores"] = req.CPUCores }
	if req.RAMGB > 0 { updates["ram_gb"] = req.RAMGB }
	if req.DiskGB > 0 { updates["disk_gb"] = req.DiskGB }
	if req.PricePerHour > 0 { updates["price_per_hour"] = FromUSD(req.PricePerHour) }
	if req.Description != "" { updates["description"] = req.Description }
	if req.IsAvailable != nil { updates["is_available"] = *req.IsAvailable }

	if err := h.specService.UpdateSpec(c.Request.Context(), req.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteSpec 删除规格
func (h *SpecHandler) DeleteSpec(c *gin.Context) {
	var req struct {
		ID int64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.specService.DeleteSpec(c.Request.Context(), req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// EstimateCost 估算成本
func (h *SpecHandler) EstimateCost(c *gin.Context) {
	var req struct {
		SpecName string  `json:"spec_name" binding:"required"`
		Hours    float64 `json:"hours" binding:"required"`
		Replicas int     `json:"replicas" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cost, err := h.specService.EstimateCost(c.Request.Context(), req.SpecName, req.Hours, req.Replicas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"spec_name": req.SpecName, "hours": req.Hours, "replicas": req.Replicas,
		"estimated_cost": cost, "currency": "USD",
	})
}
