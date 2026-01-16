package service

import (
	"context"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
)

type ClusterService struct {
	repo *mysql.ClusterRepo
}

func NewClusterService(repo *mysql.ClusterRepo) *ClusterService {
	return &ClusterService{repo: repo}
}

func (s *ClusterService) CreateCluster(ctx context.Context, clusterID, clusterName, region, apiEndpoint, apiKey, status string, priority int) error {
	return s.repo.Create(ctx, &model.Cluster{
		ClusterID: clusterID, ClusterName: clusterName, Region: region,
		APIEndpoint: apiEndpoint, APIKey: apiKey, Status: status, Priority: priority,
	})
}

func (s *ClusterService) UpdateCluster(ctx context.Context, clusterID, clusterName, region, apiEndpoint, apiKey, status string, priority int) error {
	updates := map[string]interface{}{}
	if clusterName != "" {
		updates["cluster_name"] = clusterName
	}
	if region != "" {
		updates["region"] = region
	}
	if apiEndpoint != "" {
		updates["api_endpoint"] = apiEndpoint
	}
	if apiKey != "" {
		updates["api_key_hash"] = apiKey
	}
	if status != "" {
		updates["status"] = status
	}
	if priority > 0 {
		updates["priority"] = priority
	}
	return s.repo.Update(ctx, clusterID, updates)
}

func (s *ClusterService) DeleteCluster(ctx context.Context, clusterID string) error {
	return s.repo.Delete(ctx, clusterID)
}

func (s *ClusterService) GetCluster(ctx context.Context, clusterID string) (*model.Cluster, error) {
	return s.repo.GetByID(ctx, clusterID)
}

func (s *ClusterService) ListClusters(ctx context.Context) ([]model.Cluster, error) {
	return s.repo.List(ctx)
}

func (s *ClusterService) ListActiveClusters(ctx context.Context) ([]model.Cluster, error) {
	return s.repo.ListActive(ctx)
}

func (s *ClusterService) GetClusterSpecs(ctx context.Context, clusterID string) ([]model.ClusterSpec, error) {
	return s.repo.ListSpecsByCluster(ctx, clusterID)
}

func (s *ClusterService) CreateClusterSpec(ctx context.Context, spec *model.ClusterSpec) error {
	return s.repo.CreateSpec(ctx, spec)
}

func (s *ClusterService) UpdateClusterSpec(ctx context.Context, id int64, updates map[string]interface{}) error {
	return s.repo.UpdateSpec(ctx, id, updates)
}

func (s *ClusterService) DeleteClusterSpec(ctx context.Context, id int64) error {
	return s.repo.DeleteSpec(ctx, id)
}

func (s *ClusterService) MarkOfflineClusters(ctx context.Context, timeout time.Duration) error {
	return s.repo.MarkOffline(ctx, timeout)
}
