package waverless

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client Waverless API 客户端
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient 创建 Waverless 客户端
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateEndpointRequest 创建 Endpoint 请求
type CreateEndpointRequest struct {
	Endpoint          string            `json:"endpoint"`
	SpecName          string            `json:"specName"`
	Image             string            `json:"image"`
	Replicas          int               `json:"replicas,omitempty"`
	TaskTimeout       int               `json:"taskTimeout,omitempty"`
	MaxPendingTasks   int               `json:"maxPendingTasks,omitempty"`
	ShmSize           string            `json:"shmSize,omitempty"`
	EnablePtrace      bool              `json:"enablePtrace,omitempty"`
	MinReplicas       int               `json:"minReplicas,omitempty"`
	MaxReplicas       int               `json:"maxReplicas,omitempty"`
	ScaleUpThreshold  int               `json:"scaleUpThreshold,omitempty"`
	ScaleDownIdleTime int               `json:"scaleDownIdleTime,omitempty"`
	ScaleUpCooldown   int               `json:"scaleUpCooldown,omitempty"`
	ScaleDownCooldown int               `json:"scaleDownCooldown,omitempty"`
	Priority          int               `json:"priority,omitempty"`
	EnableDynamicPrio bool              `json:"enableDynamicPrio,omitempty"`
	HighLoadThreshold int               `json:"highLoadThreshold,omitempty"`
	PriorityBoost     int               `json:"priorityBoost,omitempty"`
	Env               map[string]string `json:"env,omitempty"`
}

// SubmitTaskRequest 提交任务请求
type SubmitTaskRequest struct {
	Input map[string]interface{} `json:"input"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID            string                 `json:"id"`
	Status        string                 `json:"status"`
	Endpoint      string                 `json:"endpoint,omitempty"`
	WorkerID      string                 `json:"workerId,omitempty"`
	DelayTime     int64                  `json:"delayTime,omitempty"`
	ExecutionTime int64                  `json:"executionTime,omitempty"`
	CreatedAt     string                 `json:"createdAt,omitempty"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Output        map[string]interface{} `json:"output,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

// CreateEndpoint 创建 Endpoint
func (c *Client) CreateEndpoint(ctx context.Context, req *CreateEndpointRequest) error {
	return c.doRequest(ctx, "POST", "/api/v1/endpoints", req, nil)
}

// GetEndpoint 获取 Endpoint 详情
func (c *Client) GetEndpoint(ctx context.Context, name string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/endpoints/%s", name), nil, &resp)
	return resp, err
}

// DeleteEndpoint 删除 Endpoint
func (c *Client) DeleteEndpoint(ctx context.Context, name string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/endpoints/%s", name), nil, nil)
}

// UpdateEndpointDeployment 更新 Endpoint 部署 (replicas, image, env)
func (c *Client) UpdateEndpointDeployment(ctx context.Context, name string, replicas int, image string, env map[string]string) error {
	body := map[string]interface{}{}
	if replicas > 0 {
		body["replicas"] = replicas
	}
	if image != "" {
		body["image"] = image
	}
	if len(env) > 0 {
		body["env"] = env
	}
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/v1/endpoints/%s/deployment", name), body, nil)
}

// UpdateEndpointConfig 更新 Endpoint 配置
func (c *Client) UpdateEndpointConfig(ctx context.Context, name string, config map[string]interface{}) error {
	return c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/autoscaler/endpoints/%s", name), config, nil)
}

// GetEndpointWorkers 获取 Endpoint Worker 列表 (只返回活跃的)
func (c *Client) GetEndpointWorkers(ctx context.Context, name string) ([]map[string]interface{}, error) {
	var resp []map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/endpoints/%s/workers", name), nil, &resp)
	return resp, err
}

// GetWorker 获取单个 Worker 详情 (包括 OFFLINE)
func (c *Client) GetWorker(ctx context.Context, workerID string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/workers/%s", workerID), nil, &resp)
	return resp, err
}

// GetWorkerLogs 获取 Worker 日志
func (c *Client) GetWorkerLogs(ctx context.Context, endpoint, podName string, lines int) (string, error) {
	url := fmt.Sprintf("%s/api/v1/endpoints/%s/logs?lines=%d&pod_name=%s", c.baseURL, endpoint, lines, podName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

// GetWorkerTasks 获取 Worker 任务列表 (兼容旧调用)
func (c *Client) GetWorkerTasks(ctx context.Context, endpoint, workerID string, limit int) (map[string]interface{}, error) {
	return c.GetTasks(ctx, endpoint, workerID, "", "", fmt.Sprintf("%d", limit), "0")
}

// GetTasks 获取任务列表
func (c *Client) GetTasks(ctx context.Context, endpoint, workerID, status, taskID, limit, offset string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	url := fmt.Sprintf("/v1/tasks?endpoint=%s&limit=%s&offset=%s", endpoint, limit, offset)
	if workerID != "" {
		url += "&worker_id=" + workerID
	}
	if status != "" {
		url += "&status=" + status
	}
	if taskID != "" {
		url += "&task_id=" + taskID
	}
	err := c.doRequest(ctx, "GET", url, nil, &resp)
	return resp, err
}

// GetEndpointMetrics 获取 Endpoint 实时指标
func (c *Client) GetEndpointMetrics(ctx context.Context, name string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/%s/metrics/realtime", name), nil, &resp)
	return resp, err
}

// GetEndpointStatistics 获取 Endpoint 统计信息
func (c *Client) GetEndpointStatistics(ctx context.Context, name string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/statistics/endpoints/%s", name), nil, &resp)
	return resp, err
}

// GetTasksOverview 获取任务总览统计
func (c *Client) GetTasksOverview(ctx context.Context) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", "/api/v1/statistics/overview", nil, &resp)
	return resp, err
}

// GetEndpointStats 获取 Endpoint 统计数据 (透传)
func (c *Client) GetEndpointStats(ctx context.Context, name, from, to string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v1/%s/metrics/stats", name)
	params := []string{}
	if from != "" {
		params = append(params, "from="+from)
	}
	if to != "" {
		params = append(params, "to="+to)
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", path, nil, &resp)
	return resp, err
}

// SubmitTask 提交任务
func (c *Client) SubmitTask(ctx context.Context, endpoint string, input map[string]interface{}) (*TaskResponse, error) {
	req := &SubmitTaskRequest{Input: input}
	var resp TaskResponse
	err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/%s/run", endpoint), req, &resp)
	return &resp, err
}

// SubmitTaskSync 同步提交任务
func (c *Client) SubmitTaskSync(ctx context.Context, endpoint string, input map[string]interface{}) (*TaskResponse, error) {
	req := &SubmitTaskRequest{Input: input}
	var resp TaskResponse
	err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/%s/runsync", endpoint), req, &resp)
	return &resp, err
}

// GetTaskStatus 获取任务状态
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*TaskResponse, error) {
	var resp TaskResponse
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/status/%s", taskID), nil, &resp)
	return &resp, err
}

// CancelTask 取消任务
func (c *Client) CancelTask(ctx context.Context, taskID string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/v1/cancel/%s", taskID), nil, nil)
}

// GetTaskTimeline 获取任务时间线
func (c *Client) GetTaskTimeline(ctx context.Context, taskID string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/tasks/%s/timeline", taskID), nil, &resp)
	return resp, err
}

// GetTaskExecutionHistory 获取任务执行历史
func (c *Client) GetTaskExecutionHistory(ctx context.Context, taskID string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/tasks/%s/execution-history", taskID), nil, &resp)
	return resp, err
}

// GetScalingHistory 获取扩缩容历史
func (c *Client) GetScalingHistory(ctx context.Context, endpoint string, limit int) ([]map[string]interface{}, error) {
	var resp []map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/autoscaler/endpoints/%s/history?limit=%d", endpoint, limit), nil, &resp)
	return resp, err
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
