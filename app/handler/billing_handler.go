package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"

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

// GetBalance 获取余额
func (h *BillingHandler) GetBalance(c *gin.Context) {
	userID := c.GetString("user_id")

	balance, err := h.userService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":               balance.Balance,
		"credit_limit":          balance.CreditLimit,
		"currency":              balance.Currency,
		"status":                balance.Status,
		"low_balance_threshold": balance.LowBalanceThreshold,
	})
}

// GetUsage 获取使用统计
func (h *BillingHandler) GetUsage(c *gin.Context) {
	userID := c.GetString("user_id")

	// 默认查询最近 30 天
	to := time.Now()
	from := to.AddDate(0, 0, -30)

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

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetRechargeRecords 获取充值记录
func (h *BillingHandler) GetRechargeRecords(c *gin.Context) {
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

	records, total, err := h.userService.GetRechargeRecords(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
