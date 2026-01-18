package model

import (
	"time"

	"gorm.io/datatypes"
)

// Worker Portal 本地缓存的 Worker 信息
type Worker struct {
	ID                   int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkerID             string         `gorm:"uniqueIndex;not null" json:"worker_id"`
	EndpointID           int64          `gorm:"index;not null" json:"endpoint_id"`
	ClusterID            string         `gorm:"index;not null" json:"cluster_id"`
	UserID               string         `gorm:"index;not null" json:"user_id"`
	PodName              string         `json:"pod_name"`
	Status               string         `gorm:"default:STARTING" json:"status"` // STARTING, ONLINE, BUSY, DRAINING, OFFLINE
	PodCreatedAt         *time.Time     `json:"pod_created_at"`
	PodStartedAt         *time.Time     `json:"pod_started_at"`
	PodReadyAt           *time.Time     `json:"pod_ready_at"`
	PodTerminatedAt      *time.Time     `json:"pod_terminated_at"`
	ColdStartDurationMs  *int64         `json:"cold_start_duration_ms"`
	CurrentJobs          int            `gorm:"default:0" json:"current_jobs"`
	TotalTasksCompleted  int64          `gorm:"default:0" json:"total_tasks_completed"`
	TotalTasksFailed     int64          `gorm:"default:0" json:"total_tasks_failed"`
	TotalExecutionTimeMs int64          `gorm:"default:0" json:"total_execution_time_ms"`
	LastTaskTime         *time.Time     `json:"last_task_time"`
	LastHeartbeat        *time.Time     `json:"last_heartbeat"`
	BillingStatus        string         `gorm:"default:pending" json:"billing_status"` // pending, active, final_billed
	LastBilledAt         *time.Time     `json:"last_billed_at"`
	TotalBilledSeconds   int            `gorm:"default:0" json:"total_billed_seconds"`
	TotalBilledAmount    int64          `gorm:"type:bigint;default:0" json:"total_billed_amount"`
	LastSyncedAt         *time.Time     `json:"last_synced_at"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

func (Worker) TableName() string { return "workers" }

// TaskRouting 任务路由记录
type TaskRouting struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID          string         `gorm:"uniqueIndex;not null" json:"task_id"`
	UserID          string         `gorm:"index;not null" json:"user_id"`
	OrgID           string         `json:"org_id"`
	EndpointID      int64          `gorm:"index;not null" json:"endpoint_id"`
	ClusterID       string         `gorm:"not null" json:"cluster_id"`
	Input           datatypes.JSON `json:"input"`
	WorkerID        string         `json:"worker_id"`
	Status          string         `gorm:"default:PENDING;index" json:"status"`
	SubmittedAt     time.Time      `json:"submitted_at"`
	CreatedAt       *time.Time     `json:"created_at"`
	CompletedAt     *time.Time     `json:"completed_at"`
	ExecutionTimeMs int64          `gorm:"default:0" json:"execution_time_ms"`
}

func (TaskRouting) TableName() string { return "task_routing" }

// EndpointMinuteStat 分钟级统计
type EndpointMinuteStat struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EndpointID        int64     `gorm:"uniqueIndex:uk_endpoint_minute;not null" json:"endpoint_id"`
	StatMinute        time.Time `gorm:"uniqueIndex:uk_endpoint_minute;index" json:"stat_minute"`
	ActiveWorkers     int       `gorm:"default:0" json:"active_workers"`
	IdleWorkers       int       `gorm:"default:0" json:"idle_workers"`
	TasksSubmitted    int       `gorm:"default:0" json:"tasks_submitted"`
	TasksCompleted    int       `gorm:"default:0" json:"tasks_completed"`
	TasksFailed       int       `gorm:"default:0" json:"tasks_failed"`
	TasksTimeout      int       `gorm:"default:0" json:"tasks_timeout"`
	AvgQueueWaitMs    float64   `gorm:"type:decimal(10,2);default:0" json:"avg_queue_wait_ms"`
	AvgExecutionMs    float64   `gorm:"type:decimal(10,2);default:0" json:"avg_execution_ms"`
	P95ExecutionMs    float64   `gorm:"type:decimal(10,2);default:0" json:"p95_execution_ms"`
	WorkersCreated    int       `gorm:"default:0" json:"workers_created"`
	WorkersTerminated int       `gorm:"default:0" json:"workers_terminated"`
	ColdStarts        int       `gorm:"default:0" json:"cold_starts"`
	AvgColdStartMs    float64   `gorm:"type:decimal(10,2);default:0" json:"avg_cold_start_ms"`
	CreatedAt         time.Time `json:"created_at"`
}

func (EndpointMinuteStat) TableName() string { return "endpoint_minute_stats" }

// EndpointHourlyStat 小时级统计
type EndpointHourlyStat struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EndpointID         int64     `gorm:"uniqueIndex:uk_endpoint_hour;not null" json:"endpoint_id"`
	StatHour           time.Time `gorm:"uniqueIndex:uk_endpoint_hour;index" json:"stat_hour"`
	ActiveWorkers      int       `gorm:"default:0" json:"active_workers"`
	IdleWorkers        int       `gorm:"default:0" json:"idle_workers"`
	AvgWorkers         float64   `gorm:"type:decimal(10,2);default:0" json:"avg_workers"`
	MaxWorkers         int       `gorm:"default:0" json:"max_workers"`
	TasksSubmitted     int       `gorm:"default:0" json:"tasks_submitted"`
	TasksCompleted     int       `gorm:"default:0" json:"tasks_completed"`
	TasksFailed        int       `gorm:"default:0" json:"tasks_failed"`
	TasksTimeout       int       `gorm:"default:0" json:"tasks_timeout"`
	AvgQueueWaitMs     float64   `gorm:"type:decimal(10,2);default:0" json:"avg_queue_wait_ms"`
	AvgExecutionMs     float64   `gorm:"type:decimal(10,2);default:0" json:"avg_execution_ms"`
	P50ExecutionMs     float64   `gorm:"type:decimal(10,2);default:0" json:"p50_execution_ms"`
	P95ExecutionMs     float64   `gorm:"type:decimal(10,2);default:0" json:"p95_execution_ms"`
	WorkersCreated     int       `gorm:"default:0" json:"workers_created"`
	WorkersTerminated  int       `gorm:"default:0" json:"workers_terminated"`
	ColdStarts         int       `gorm:"default:0" json:"cold_starts"`
	AvgColdStartMs     float64   `gorm:"type:decimal(10,2);default:0" json:"avg_cold_start_ms"`
	TotalWorkerSeconds int64     `gorm:"default:0" json:"total_worker_seconds"`
	TotalCost          float64   `gorm:"type:decimal(12,4);default:0" json:"total_cost"`
	CreatedAt          time.Time `json:"created_at"`
}

func (EndpointHourlyStat) TableName() string { return "endpoint_hourly_stats" }

// EndpointDailyStat 日级统计
type EndpointDailyStat struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EndpointID        int64     `gorm:"uniqueIndex:uk_endpoint_date;not null" json:"endpoint_id"`
	StatDate          time.Time `gorm:"type:date;uniqueIndex:uk_endpoint_date;index" json:"stat_date"`
	AvgWorkers        float64   `gorm:"type:decimal(10,2);default:0" json:"avg_workers"`
	MaxWorkers        int       `gorm:"default:0" json:"max_workers"`
	PeakHour          *int      `json:"peak_hour"`
	TasksSubmitted    int       `gorm:"default:0" json:"tasks_submitted"`
	TasksCompleted    int       `gorm:"default:0" json:"tasks_completed"`
	TasksFailed       int       `gorm:"default:0" json:"tasks_failed"`
	TasksTimeout      int       `gorm:"default:0" json:"tasks_timeout"`
	SuccessRate       float64   `gorm:"type:decimal(5,2);default:0" json:"success_rate"`
	AvgQueueWaitMs    float64   `gorm:"type:decimal(10,2);default:0" json:"avg_queue_wait_ms"`
	AvgExecutionMs    float64   `gorm:"type:decimal(10,2);default:0" json:"avg_execution_ms"`
	P50ExecutionMs    float64   `gorm:"type:decimal(10,2);default:0" json:"p50_execution_ms"`
	P95ExecutionMs    float64   `gorm:"type:decimal(10,2);default:0" json:"p95_execution_ms"`
	WorkersCreated    int       `gorm:"default:0" json:"workers_created"`
	WorkersTerminated int       `gorm:"default:0" json:"workers_terminated"`
	ColdStarts        int       `gorm:"default:0" json:"cold_starts"`
	AvgColdStartMs    float64   `gorm:"type:decimal(10,2);default:0" json:"avg_cold_start_ms"`
	TotalWorkerHours  float64   `gorm:"type:decimal(10,2);default:0" json:"total_worker_hours"`
	TotalCost         float64   `gorm:"type:decimal(12,4);default:0" json:"total_cost"`
	CreatedAt         time.Time `json:"created_at"`
}

func (EndpointDailyStat) TableName() string { return "endpoint_daily_stats" }
