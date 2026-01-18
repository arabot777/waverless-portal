package wavespeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/utils"
)

// BalanceResponse 余额响应
type BalanceResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		OrgID          string `json:"org_id"`
		Balance        int64  `json:"balance"`         // 微美元
		RewardBalance  int64  `json:"reward_balance"`  // 微美元
		CreditLimit    int64  `json:"credit_limit"`    // 微美元
		TotalAvailable int64  `json:"total_available"` // 微美元
	} `json:"data"`
}

// GetOrgBalance 从主站获取组织余额 (返回微美元 int64) - 用户请求时使用 cookie
func GetOrgBalance(ctx context.Context, orgID, cookie string) (int64, error) {
	mainSiteURL := config.GlobalConfig.MainSite.URL
	endpoint := fmt.Sprintf("/center/default/api/v1/organization_with_credit/%s", orgID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mainSiteURL+endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	cookieName := config.GetCookieName()
	req.AddCookie(&http.Cookie{Name: cookieName, Value: cookie})
	req.Header.Set("Accept", "application/json")

	return doBalanceRequest(req)
}

// GetOrgBalanceInternal 内部服务调用获取余额 - 后台任务使用
func GetOrgBalanceInternal(ctx context.Context, orgID string) (int64, error) {
	apiURL := config.GlobalConfig.MainSite.APIURL
	if apiURL == "" {
		apiURL = config.GlobalConfig.MainSite.URL
	}
	endpoint := fmt.Sprintf("/api/internal/v1/org/%s/balance", orgID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	utils.SignRequest(req, "waverless-portal", config.GlobalConfig.MainSite.InternalServiceKey)
	req.Header.Set("Accept", "application/json")

	return doBalanceRequest(req)
}

func doBalanceRequest(req *http.Request) (int64, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("Balance API returned status %d: %s", resp.StatusCode, string(body))
		return 0, fmt.Errorf("balance API failed with status %d", resp.StatusCode)
	}

	var response BalanceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Code != 200 {
		return 0, fmt.Errorf("balance API error: %s", response.Message)
	}

	// 优先用 total_available，否则计算
	if response.Data.TotalAvailable > 0 {
		return response.Data.TotalAvailable, nil
	}
	return response.Data.Balance + response.Data.RewardBalance + response.Data.CreditLimit, nil
}
