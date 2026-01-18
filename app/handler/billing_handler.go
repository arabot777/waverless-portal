package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/wavespeed"

	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	billingService *service.BillingService
	userService    *service.UserService
}

func NewBillingHandler(billingService *service.BillingService, userService *service.UserService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
		userService:    userService,
	}
}

// GetBalance 获取余额 (从主站获取)
func (h *BillingHandler) GetBalance(c *gin.Context) {
	orgID := c.GetString("org_id")
	cookieName := config.GetCookieName()
	cookie, _ := c.Cookie(cookieName)

	balance, err := wavespeed.GetOrgBalance(c.Request.Context(), orgID, cookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":  ToUSD(balance),
		"currency": "USD",
	})
}

// GetUsage 获取使用统计
func (h *BillingHandler) GetUsage(c *gin.Context) {
	userID := c.GetString("user_id")

	// 默认查询最近 7 天
	to := time.Now()
	from := to.AddDate(0, 0, -7)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	stats, err := h.billingService.GetUsageStats(c.Request.Context(), userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换金额为 USD
	stats["total_amount"] = ToUSD(stats["total_amount"].(int64))
	c.JSON(http.StatusOK, stats)
}

// GetWorkerRecords 获取 Worker 计费记录
func (h *BillingHandler) GetWorkerRecords(c *gin.Context) {
	userID := c.GetString("user_id")

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	records, total, err := h.billingService.GetWorkerBillingRecords(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换金额为 USD
	result := make([]map[string]interface{}, len(records))
	for i, r := range records {
		result[i] = map[string]interface{}{
			"id":                   r.ID,
			"worker_id":            r.WorkerID,
			"endpoint_id":          r.EndpointID,
			"endpoint_name":        r.EndpointName,
			"spec_name":            r.SpecName,
			"billing_period_start": r.BillingPeriodStart,
			"billing_period_end":   r.BillingPeriodEnd,
			"duration_seconds":     r.DurationSeconds,
			"price_per_hour":       ToUSD(r.PricePerHour),
			"amount":               ToUSD(r.Amount),
			"status":               r.Status,
			"created_at":           r.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"records": result,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
