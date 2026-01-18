package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
)

type Manager struct {
	clusterService *service.ClusterService
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

func NewManager(billingService *service.BillingService, clusterService *service.ClusterService) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		clusterService: clusterService,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (m *Manager) Start() {
	// 集群健康检查任务
	m.wg.Add(1)
	go m.runClusterHealthCheck()
}

func (m *Manager) Stop() {
	m.cancel()
}

func (m *Manager) Wait() {
	m.wg.Wait()
}

func (m *Manager) runClusterHealthCheck() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	logger.Infof("Cluster health check job started")

	for {
		select {
		case <-m.ctx.Done():
			logger.Infof("Cluster health check job stopped")
			return
		case <-ticker.C:
			if err := m.clusterService.MarkOfflineClusters(m.ctx, 2*time.Minute); err != nil {
				logger.Errorf("Cluster health check error: %v", err)
			}
		}
	}
}
