package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"github.com/wavespeedai/waverless-portal/pkg/waverless"
	"go.uber.org/zap"
)

type EndpointService struct {
	repo             *mysql.EndpointRepo
	clusterRepo      *mysql.ClusterRepo
	specRepo         *mysql.SpecRepo
	clusterService   *ClusterService
	waverlessClients sync.Map
}

func NewEndpointService(repo *mysql.EndpointRepo, clusterRepo *mysql.ClusterRepo, specRepo *mysql.SpecRepo, clusterService *ClusterService) *EndpointService {
	return &EndpointService{repo: repo, clusterRepo: clusterRepo, specRepo: specRepo, clusterService: clusterService}
}

type CreateEndpointRequest struct {
	LogicalName  string            `json:"logical_name"`
	SpecName     string            `json:"spec_name"`
	Image        string            `json:"image"`
	Replicas     int               `json:"replicas"`
	MinReplicas  int               `json:"min_replicas"`
	MaxReplicas  int               `json:"max_replicas"`
	TaskTimeout  int               `json:"task_timeout"`
	Env          map[string]string `json:"env"`
	PreferRegion string            `json:"prefer_region"`
}

type ClusterCandidate struct {
	Cluster     *model.Cluster
	ClusterSpec *model.ClusterSpec
	SpecPricing *model.SpecPricing
	Score       float64
}

func (s *EndpointService) Create(ctx context.Context, userID, orgID string, req *CreateEndpointRequest) (*model.UserEndpoint, error) {
	_, err := s.repo.GetByLogicalName(ctx, userID, req.LogicalName)
	if err == nil {
		return nil, errors.New("endpoint already exists")
	}

	candidate, err := s.selectBestCluster(ctx, req.SpecName, req.PreferRegion)
	if err != nil {
		return nil, err
	}

	sp := candidate.SpecPricing
	// physicalName := strings.ToLower(fmt.Sprintf("user-%s-%s", userID[:8], req.LogicalName))
	physicalName := strings.ToLower(req.LogicalName)
	endpoint := &model.UserEndpoint{
		UserID: userID, OrgID: orgID, LogicalName: req.LogicalName, PhysicalName: physicalName,
		SpecName: req.SpecName, SpecType: sp.SpecType, GPUType: sp.GPUType,
		GPUCount: sp.GPUCount, CPUCores: sp.CPUCores, RAMGB: sp.RAMGB,
		ClusterID: candidate.Cluster.ClusterID, Replicas: req.Replicas, MinReplicas: req.MinReplicas, MaxReplicas: req.MaxReplicas,
		Image: req.Image, TaskTimeout: req.TaskTimeout, PricePerHour: sp.PricePerHour,
		Currency: "USD", PreferRegion: req.PreferRegion, Status: "deploying",
	}

	client := s.getWaverlessClient(candidate.Cluster)
	replicas := req.Replicas
	if replicas == 0 {
		replicas = req.MinReplicas
	}
	if err := client.CreateEndpoint(ctx, &waverless.CreateEndpointRequest{
		Endpoint: physicalName, SpecName: candidate.ClusterSpec.ClusterSpecName, Image: req.Image,
		Replicas: replicas, MinReplicas: req.MinReplicas, MaxReplicas: req.MaxReplicas,
		TaskTimeout: req.TaskTimeout, Env: req.Env,
	}); err != nil {
		logger.ErrorCtx(ctx, "failed to create endpoint in waverless", zap.Error(err))
		return nil, fmt.Errorf("failed to create endpoint in waverless: %w", err)
	}

	if err := s.repo.Create(ctx, endpoint); err != nil {
		client.DeleteEndpoint(ctx, physicalName)
		return nil, err
	}
	s.repo.Update(ctx, endpoint.ID, map[string]interface{}{"status": "running"})
	endpoint.Status = "running"
	return endpoint, nil
}

func (s *EndpointService) selectBestCluster(ctx context.Context, specName, preferRegion string) (*ClusterCandidate, error) {
	// 获取规格定价信息
	specPricing, err := s.specRepo.GetByName(ctx, specName)
	if err != nil {
		return nil, errors.New("spec not found")
	}

	// 获取支持该规格的集群
	clusterSpecs, err := s.clusterRepo.ListAvailableSpecsByName(ctx, specName)
	if err != nil || len(clusterSpecs) == 0 {
		return nil, errors.New("no available cluster for this spec")
	}

	var candidates []ClusterCandidate
	for _, cs := range clusterSpecs {
		cluster, err := s.clusterRepo.GetByID(ctx, cs.ClusterID)
		if err != nil || cluster.Status != "active" {
			continue
		}
		candidates = append(candidates, ClusterCandidate{
			Cluster: cluster, ClusterSpec: &cs, SpecPricing: specPricing,
		})
	}
	if len(candidates) == 0 {
		return nil, errors.New("no active cluster available")
	}

	// 评分：可用性 50%，区域偏好 30%，优先级 20%
	for i := range candidates {
		c := &candidates[i]
		availScore := float64(c.ClusterSpec.AvailableCapacity) / float64(c.ClusterSpec.TotalCapacity+1)
		regionScore := 0.0
		if preferRegion != "" && c.Cluster.Region == preferRegion {
			regionScore = 1.0
		}
		priorityScore := float64(c.Cluster.Priority) / 100.0
		c.Score = availScore*0.5 + regionScore*0.3 + priorityScore*0.2
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })
	return &candidates[0], nil
}

func (s *EndpointService) getWaverlessClient(cluster *model.Cluster) *waverless.Client {
	if client, ok := s.waverlessClients.Load(cluster.ClusterID); ok {
		return client.(*waverless.Client)
	}
	client := waverless.NewClient(cluster.APIEndpoint, cluster.APIKey)
	s.waverlessClients.Store(cluster.ClusterID, client)
	return client
}

func (s *EndpointService) GetWaverlessClient(cluster *model.Cluster) *waverless.Client {
	return s.getWaverlessClient(cluster)
}

func (s *EndpointService) GetByLogicalName(ctx context.Context, userID, logicalName string) (*model.UserEndpoint, error) {
	return s.repo.GetByLogicalName(ctx, userID, logicalName)
}

func (s *EndpointService) GetEndpointDetail(ctx context.Context, endpoint *model.UserEndpoint) (map[string]interface{}, error) {
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return nil, err
	}
	return s.getWaverlessClient(cluster).GetEndpoint(ctx, endpoint.PhysicalName)
}

func (s *EndpointService) List(ctx context.Context, userID string) ([]model.UserEndpoint, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *EndpointService) ListAll(ctx context.Context) ([]model.UserEndpoint, error) {
	return s.repo.ListAll(ctx)
}

func (s *EndpointService) Delete(ctx context.Context, userID, logicalName string) error {
	endpoint, err := s.repo.GetByLogicalName(ctx, userID, logicalName)
	if err != nil {
		return err
	}
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return err
	}
	if err := s.getWaverlessClient(cluster).DeleteEndpoint(ctx, endpoint.PhysicalName); err != nil {
		return fmt.Errorf("failed to delete endpoint in waverless: %w", err)
	}
	return s.repo.SoftDelete(ctx, endpoint.ID)
}

// ScaleEndpoint 调整 Endpoint 副本数
func (s *EndpointService) ScaleEndpoint(ctx context.Context, endpoint *model.UserEndpoint, replicas int) error {
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return err
	}
	if err := s.getWaverlessClient(cluster).UpdateEndpointDeployment(ctx, endpoint.PhysicalName, replicas, "", nil); err != nil {
		return err
	}
	return s.repo.Update(ctx, endpoint.ID, map[string]interface{}{"replicas": replicas})
}

func (s *EndpointService) UpdateDeployment(ctx context.Context, userID, logicalName string, replicas int, image string, env map[string]string) error {
	endpoint, err := s.repo.GetByLogicalName(ctx, userID, logicalName)
	if err != nil {
		return err
	}
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return err
	}
	if err := s.getWaverlessClient(cluster).UpdateEndpointDeployment(ctx, endpoint.PhysicalName, replicas, image, env); err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if image != "" {
		updates["image"] = image
	}
	if len(updates) > 0 {
		return s.repo.Update(ctx, endpoint.ID, updates)
	}
	return nil
}

// UpdateConfig 更新 Endpoint 配置
func (s *EndpointService) UpdateConfig(ctx context.Context, userID, logicalName string, config map[string]interface{}) error {
	endpoint, err := s.repo.GetByLogicalName(ctx, userID, logicalName)
	if err != nil {
		return err
	}
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return err
	}
	if err := s.getWaverlessClient(cluster).UpdateEndpointConfig(ctx, endpoint.PhysicalName, config); err != nil {
		return err
	}
	// 更新本地数据库
	updates := map[string]interface{}{}
	if v, ok := config["minReplicas"]; ok {
		updates["min_replicas"] = v
	}
	if v, ok := config["maxReplicas"]; ok {
		updates["max_replicas"] = v
	}
	if v, ok := config["taskTimeout"]; ok {
		updates["task_timeout"] = v
	}
	if len(updates) > 0 {
		return s.repo.Update(ctx, endpoint.ID, updates)
	}
	return nil
}

// GetSpecPrice 获取规格价格
func (s *EndpointService) GetSpecPrice(ctx context.Context, specName string) (int64, error) {
	spec, err := s.specRepo.GetByName(ctx, specName)
	if err != nil {
		return 0, err
	}
	return spec.PricePerHour, nil
}
