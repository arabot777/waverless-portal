package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wavespeedai/waverless-portal/internal/jobs"
	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
)

type MonitoringHandler struct {
	endpointService *service.EndpointService
	clusterService  *service.ClusterService
	taskRepo        *mysql.TaskRepo
	workerRepo      *mysql.WorkerRepo
	workerSyncJob   *jobs.WorkerSyncJob
}

func NewMonitoringHandler(endpointService *service.EndpointService, clusterService *service.ClusterService, taskRepo *mysql.TaskRepo, workerRepo *mysql.WorkerRepo, workerSyncJob *jobs.WorkerSyncJob) *MonitoringHandler {
	return &MonitoringHandler{
		endpointService: endpointService,
		clusterService:  clusterService,
		taskRepo:        taskRepo,
		workerRepo:      workerRepo,
		workerSyncJob:   workerSyncJob,
	}
}

func (h *MonitoringHandler) SetWorkerSyncJob(job *jobs.WorkerSyncJob) {
	h.workerSyncJob = job
}

// GetEndpointWorkers 获取 Endpoint 的 Worker 列表 (从本地 workers 表)
func (h *MonitoringHandler) GetEndpointWorkers(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")

	endpoint, err := h.endpointService.GetByLogicalName(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	// 先同步再返回
	if h.workerSyncJob != nil {
		h.workerSyncJob.SyncEndpoint(c.Request.Context(), endpoint)
	}

	workers, err := h.workerRepo.ListByEndpoint(c.Request.Context(), endpoint.ID)
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
	metrics["price_per_hour"] = ToUSD(endpoint.PricePerHour)
	metrics["endpoint_name"] = name

	c.JSON(http.StatusOK, metrics)
}

// GetEndpointStats 获取 Endpoint 统计数据 (透传 waverless)
func (h *MonitoringHandler) GetEndpointStats(c *gin.Context) {
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

	// 直接透传到 waverless
	client := h.endpointService.GetWaverlessClient(cluster)
	from := c.Query("from")
	to := c.Query("to")
	resp, err := client.GetEndpointStats(c.Request.Context(), endpoint.PhysicalName, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
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

// GetAllTasks 获取用户所有任务 (从本地 task_routing 表)
func (h *MonitoringHandler) GetAllTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tasks, total, err := h.taskRepo.List(c.Request.Context(), userID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks, "total": total})
}

// GetTasksOverview 获取任务总览统计 (从本地 task_routing 表)
func (h *MonitoringHandler) GetTasksOverview(c *gin.Context) {
	userID := c.GetString("user_id")

	overview, err := h.taskRepo.GetOverview(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"completed": 0, "in_progress": 0, "pending": 0, "failed": 0})
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

// GetTaskTimeline 获取任务时间线
func (h *MonitoringHandler) GetTaskTimeline(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("task_id")

	// 从本地 task_routing 获取任务信息
	task, err := h.taskRepo.GetByTaskID(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), task.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	timeline, err := client.GetTaskTimeline(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timeline)
}

// GetTaskExecutionHistory 获取任务执行历史
func (h *MonitoringHandler) GetTaskExecutionHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("task_id")

	task, err := h.taskRepo.GetByTaskID(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), task.ClusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := h.endpointService.GetWaverlessClient(cluster)
	history, err := client.GetTaskExecutionHistory(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetScalingHistory 获取扩缩容历史
func (h *MonitoringHandler) GetScalingHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	name := c.Param("name")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

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
	history, err := client.GetScalingHistory(c.Request.Context(), endpoint.PhysicalName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
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
