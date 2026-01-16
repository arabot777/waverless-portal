package service

import (
	"context"
	"fmt"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"github.com/wavespeedai/waverless-portal/pkg/waverless"
)

type TaskService struct {
	billingRepo     *mysql.BillingRepo
	endpointService *EndpointService
	clusterService  *ClusterService
}

func NewTaskService(billingRepo *mysql.BillingRepo, endpointService *EndpointService, clusterService *ClusterService) *TaskService {
	return &TaskService{billingRepo: billingRepo, endpointService: endpointService, clusterService: clusterService}
}

func (s *TaskService) SubmitTask(ctx context.Context, userID, orgID, logicalName string, input map[string]interface{}) (*waverless.TaskResponse, error) {
	endpoint, err := s.endpointService.GetByLogicalName(ctx, userID, logicalName)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return nil, err
	}
	resp, err := s.endpointService.GetWaverlessClient(cluster).SubmitTask(ctx, endpoint.PhysicalName, input)
	if err != nil {
		return nil, err
	}
	s.billingRepo.CreateTaskRouting(ctx, &model.TaskRouting{
		TaskID: resp.ID, UserID: userID, OrgID: orgID, EndpointID: endpoint.ID, ClusterID: endpoint.ClusterID,
	})
	return resp, nil
}

func (s *TaskService) SubmitTaskSync(ctx context.Context, userID, orgID, logicalName string, input map[string]interface{}) (*waverless.TaskResponse, error) {
	endpoint, err := s.endpointService.GetByLogicalName(ctx, userID, logicalName)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return nil, err
	}
	resp, err := s.endpointService.GetWaverlessClient(cluster).SubmitTaskSync(ctx, endpoint.PhysicalName, input)
	if err != nil {
		return nil, err
	}
	s.billingRepo.CreateTaskRouting(ctx, &model.TaskRouting{
		TaskID: resp.ID, UserID: userID, OrgID: orgID, EndpointID: endpoint.ID, ClusterID: endpoint.ClusterID,
	})
	return resp, nil
}

func (s *TaskService) GetTaskStatus(ctx context.Context, userID, taskID string) (*waverless.TaskResponse, error) {
	routing, err := s.billingRepo.GetTaskRouting(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, routing.ClusterID)
	if err != nil {
		return nil, err
	}
	return s.endpointService.GetWaverlessClient(cluster).GetTaskStatus(ctx, taskID)
}

func (s *TaskService) CancelTask(ctx context.Context, userID, taskID string) error {
	routing, err := s.billingRepo.GetTaskRouting(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, routing.ClusterID)
	if err != nil {
		return err
	}
	return s.endpointService.GetWaverlessClient(cluster).CancelTask(ctx, taskID)
}
