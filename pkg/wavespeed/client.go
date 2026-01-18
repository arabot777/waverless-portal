package wavespeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/utils"
)

var envConfig *EnvConfig

type EnvConfig struct {
	MainSiteURL string
	CookieName  string
}

func GetEnvConfig() *EnvConfig {
	if envConfig == nil {
		env := os.Getenv("ENV")
		isProd := env == "production"
		envConfig = &EnvConfig{}
		if isProd {
			envConfig.MainSiteURL = "https://wavespeed.ai"
			envConfig.CookieName = "token"
		} else {
			envConfig.MainSiteURL = "https://tropical.wavespeed.ai"
			envConfig.CookieName = "test_token"
		}
	}
	return envConfig
}

func GetMainSiteURL() string { return GetEnvConfig().MainSiteURL }
func GetCookieName() string  { return GetEnvConfig().CookieName }

// User 主站用户信息
type User struct {
	UserID      string   `json:"user_id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	AvatarURL   string   `json:"avatar_url"`
	Permissions []string `json:"permissions"`
}

type UserResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    User   `json:"data"`
}

// GetUserInfo 从主站获取用户信息
func GetUserInfo(ctx context.Context, cookieValue string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GetMainSiteURL()+"/center/default/api/v1/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", GetCookieName()+"="+cookieValue)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var response UserResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("main site error: %s", response.Message)
	}

	return &response.Data, nil
}


// APIKeyInfo API Key 验证返回的信息
type APIKeyInfo struct {
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
	Email  string `json:"email"`
}

type APIKeyResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    APIKeyInfo `json:"data"`
}

// ValidateAPIKey 验证 API Key 并返回用户信息
func ValidateAPIKey(ctx context.Context, apiKey string) (*APIKeyInfo, error) {
	apiURL := config.GlobalConfig.MainSite.APIURL
	if apiURL == "" {
		apiURL = config.GlobalConfig.MainSite.URL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/api/internal/v1/apikey/validate", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	// 添加内部服务签名
	utils.SignRequest(req, "waverless-portal", config.GlobalConfig.MainSite.InternalServiceKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid api key")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var response APIKeyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("api key validation failed: %s", response.Message)
	}

	return &response.Data, nil
}
