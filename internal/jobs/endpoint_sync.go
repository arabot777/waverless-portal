package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
)

// EndpointSyncService 后台同步 endpoint 状态
type EndpointSyncService struct {
	endpointRepo    *mysql.EndpointRepo
	clusterService  *service.ClusterService
	endpointService *service.EndpointService
	interval        time.Duration
	stopCh          chan struct{}
}

func NewEndpointSyncService(
	endpointRepo *mysql.EndpointRepo,
	clusterService *service.ClusterService,
	endpointService *service.EndpointService,
) *EndpointSyncService {
	return &EndpointSyncService{
		endpointRepo:    endpointRepo,
		clusterService:  clusterService,
		endpointService: endpointService,
		interval:        10 * time.Second,
		stopCh:          make(chan struct{}),
	}
}

func (s *EndpointSyncService) Start() {
	go s.run()
}

func (s *EndpointSyncService) Stop() {
	close(s.stopCh)
}

func (s *EndpointSyncService) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 启动时立即同步一次
	s.syncAll()

	for {
		select {
		case <-ticker.C:
			s.syncAll()
		case <-s.stopCh:
			return
		}
	}
}

func (s *EndpointSyncService) syncAll() {
	ctx := context.Background()
	endpoints, err := s.endpointRepo.ListAll(ctx)
	if err != nil {
		logger.Errorf("failed to list endpoints for sync: %v", err)
		return
	}

	// 按集群分组
	clusterEndpoints := make(map[string][]struct {
		ID           int64
		PhysicalName string
	})
	for _, ep := range endpoints {
		clusterEndpoints[ep.ClusterID] = append(clusterEndpoints[ep.ClusterID], struct {
			ID           int64
			PhysicalName string
		}{ep.ID, ep.PhysicalName})
	}

	// 并发获取每个集群的 endpoint 状态
	var wg sync.WaitGroup
	for clusterID, eps := range clusterEndpoints {
		wg.Add(1)
		go func(cid string, endpoints []struct {
			ID           int64
			PhysicalName string
		}) {
			defer wg.Done()
			s.syncClusterEndpoints(ctx, cid, endpoints)
		}(clusterID, eps)
	}
	wg.Wait()
}

func (s *EndpointSyncService) syncClusterEndpoints(ctx context.Context, clusterID string, endpoints []struct {
	ID           int64
	PhysicalName string
}) {
	cluster, err := s.clusterService.GetCluster(ctx, clusterID)
	if err != nil {
		return
	}

	client := s.endpointService.GetWaverlessClient(cluster)

	for _, ep := range endpoints {
		detail, err := client.GetEndpoint(ctx, ep.PhysicalName)
		if err != nil {
			continue
		}

		// 提取状态信息
		updates := map[string]interface{}{}
		if status, ok := detail["status"].(string); ok {
			updates["status"] = status
		}
		if replicas, ok := detail["replicas"].(float64); ok {
			updates["replicas"] = int(replicas)
		}
		if readyReplicas, ok := detail["readyReplicas"].(float64); ok {
			updates["current_replicas"] = int(readyReplicas)
		}
		if image, ok := detail["image"].(string); ok {
			updates["image"] = image
		}

		if len(updates) > 0 {
			s.endpointRepo.Update(ctx, ep.ID, updates)
		}
	}
}
