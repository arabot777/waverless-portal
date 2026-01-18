package jobs

import (
	"context"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"gorm.io/gorm"
)

// MetricsSyncJob 同步 Endpoint 监控指标 (已禁用，直接透传 waverless)
type MetricsSyncJob struct{}

func NewMetricsSyncJob(db *gorm.DB, clusterService *service.ClusterService, endpointService *service.EndpointService) *MetricsSyncJob {
	return &MetricsSyncJob{}
}

func (j *MetricsSyncJob) Start(ctx context.Context) {
	// 不再同步，直接透传 waverless API
}
