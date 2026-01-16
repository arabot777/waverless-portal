package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
)

type Manager struct {
	billingService *service.BillingService
	clusterService *service.ClusterService
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

func NewManager(billingService *service.BillingService, clusterService *service.ClusterService) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		billingService: billingService,
		clusterService: clusterService,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (m *Manager) Start() {
	// 计费任务
	if config.GlobalConfig.Billing.Enabled {
		m.wg.Add(1)
		go m.runBillingJob()
	}

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

func (m *Manager) runBillingJob() {
	defer m.wg.Done()

	interval := time.Duration(config.GlobalConfig.Billing.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Infof("Billing job started, interval: %v", interval)

	for {
		select {
		case <-m.ctx.Done():
			logger.Infof("Billing job stopped")
			return
		case <-ticker.C:
			if err := m.billingService.RunBillingJob(m.ctx); err != nil {
				logger.Errorf("Billing job error: %v", err)
			}
		}
	}
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
			// 标记超过 2 分钟未心跳的集群为离线
			if err := m.clusterService.MarkOfflineClusters(m.ctx, 2*time.Minute); err != nil {
				logger.Errorf("Cluster health check error: %v", err)
			}
		}
	}
}
