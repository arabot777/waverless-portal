package handler

import (
	"net/http"

	"github.com/wavespeedai/waverless-portal/internal/service"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// SubmitTask 异步提交任务
func (h *TaskHandler) SubmitTask(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.GetString("org_id")
	endpoint := c.Param("endpoint")

	var req struct {
		Input map[string]interface{} `json:"input"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.taskService.SubmitTask(c.Request.Context(), userID, orgID, endpoint, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SubmitTaskSync 同步提交任务
func (h *TaskHandler) SubmitTaskSync(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.GetString("org_id")
	endpoint := c.Param("endpoint")

	var req struct {
		Input map[string]interface{} `json:"input"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.taskService.SubmitTaskSync(c.Request.Context(), userID, orgID, endpoint, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTaskStatus 获取任务状态
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("task_id")

	resp, err := h.taskService.GetTaskStatus(c.Request.Context(), userID, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CancelTask 取消任务
func (h *TaskHandler) CancelTask(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("task_id")

	if err := h.taskService.CancelTask(c.Request.Context(), userID, taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}
