package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wavespeedai/waverless-portal/internal/service"
)

type MonitoringHandler struct {
	endpointService *service.EndpointService
	clusterService  *service.ClusterService
}

func NewMonitoringHandler(endpointService *service.EndpointService, clusterService *service.ClusterService) *MonitoringHandler {
	return &MonitoringHandler{
		endpointService: endpointService,
		clusterService:  clusterService,
	}
}

// GetEndpointWorkers 获取 Endpoint 的 Worker 列表
func (h *MonitoringHandler) GetEndpointWorkers(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoint.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 调用 Waverless 获取 Worker 列表
	client := h.endpointService.GetWaverlessClient(cluster)
	workers, err := client.GetEndpointWorkers(c.Request.Context(), endpoint.PhysicalName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoint_name": name,
		"workers":       workers,
	})
}

// GetEndpointMetrics 获取 Endpoint 实时指标
func (h *MonitoringHandler) GetEndpointMetrics(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoint.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	metrics, err := client.GetEndpointMetrics(c.Request.Context(), endpoint.PhysicalName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 添加价格信息
	metrics["price_per_hour"] = endpoint.PricePerHour
	metrics["endpoint_name"] = name

	c.JSON(http.StatusOK, metrics)
}

// GetEndpointStats 获取 Endpoint 统计数据
func (h *MonitoringHandler) GetEndpointStats(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")
	granularity := c.DefaultQuery("granularity", "hourly") // hourly, daily
	from := c.Query("from")
	to := c.Query("to")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoint.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	stats, err := client.GetEndpointStats(c.Request.Context(), endpoint.PhysicalName, granularity, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoint_name": name,
		"granularity":   granularity,
		"stats":         stats,
	})
}


// GetEndpointStatistics 获取 Endpoint 统计信息 (透传 waverless)
func (h *MonitoringHandler) GetEndpointStatistics(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoint.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	stats, err := client.GetEndpointStatistics(c.Request.Context(), endpoint.PhysicalName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetWorkerLogs 获取 Worker 日志
func (h *MonitoringHandler) GetWorkerLogs(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")
	podName := c.Query("pod_name")
	lines := 200

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoint.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	logs, err := client.GetWorkerLogs(c.Request.Context(), endpoint.PhysicalName, podName, lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, logs)
}

// GetAllTasks 获取用户所有任务
func (h *MonitoringHandler) GetAllTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	status := c.Query("status")
	endpointFilter := c.Query("endpoint")
	taskID := c.Query("task_id")
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	endpoints, err := h.endpointService.List(c.Request.Context(), userID)
	if err != nil || len(endpoints) == 0 {
		c.JSON(http.StatusOK, gin.H{"tasks": []interface{}{}, "total": 0})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoints[0].ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var physicalNames []string
	for _, ep := range endpoints {
		if endpointFilter == "" || endpointFilter == ep.LogicalName {
			physicalNames = append(physicalNames, ep.PhysicalName)
		}
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	allTasks := []interface{}{}
	for _, pn := range physicalNames {
		tasks, err := client.GetTasks(c.Request.Context(), pn, "", status, taskID, limit, offset)
		if err == nil {
			if t, ok := tasks["tasks"].([]interface{}); ok {
				allTasks = append(allTasks, t...)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"tasks": allTasks, "total": len(allTasks)})
}

// GetTasksOverview 获取任务总览统计
func (h *MonitoringHandler) GetTasksOverview(c *gin.Context) {
	userID := c.GetString("user_id")

	endpoints, err := h.endpointService.List(c.Request.Context(), userID)
	if err != nil || len(endpoints) == 0 {
		c.JSON(http.StatusOK, gin.H{"completed": 0, "in_progress": 0, "pending": 0, "failed": 0})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoints[0].ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	overview, err := client.GetTasksOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetWorkerTasks 获取任务列表
func (h *MonitoringHandler) GetWorkerTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")
	workerID := c.Query("worker_id")
	status := c.Query("status")
	taskID := c.Query("task_id")
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoint.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	tasks, err := client.GetTasks(c.Request.Context(), endpoint.PhysicalName, workerID, status, taskID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// ExecWorker WebSocket 代理到 waverless
func (h *MonitoringHandler) ExecWorker(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")
	workerID := c.Query("worker_id")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), endpoint.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建 waverless WebSocket URL
	wsURL := cluster.APIEndpoint
	if wsURL[len(wsURL)-1] == '/' {
		wsURL = wsURL[:len(wsURL)-1]
	}
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	targetURL := fmt.Sprintf("%s/api/v1/endpoints/%s/workers/exec?worker_id=%s", wsURL, endpoint.PhysicalName, workerID)

	// 升级客户端连接
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	// 连接到 waverless
	header := http.Header{}
	if cluster.APIKey != "" {
		header.Set("Authorization", "Bearer "+cluster.APIKey)
	}
	serverConn, _, err := websocket.DefaultDialer.Dial(targetURL, header)
	if err != nil {
		clientConn.WriteMessage(websocket.TextMessage, []byte("Failed to connect to worker: "+err.Error()))
		return
	}
	defer serverConn.Close()

	// 双向转发
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			msgType, msg, err := serverConn.ReadMessage()
			if err != nil {
				return
			}
			if err := clientConn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}()

	for {
		msgType, msg, err := clientConn.ReadMessage()
		if err != nil {
			return
		}
		if err := serverConn.WriteMessage(msgType, msg); err != nil {
			return
		}
	}
}
